package service

import (
	"log"
	"bytes"
	"reflect"
	"srpc/common/sgob"
	"srpc/common/protocol"
)

// an object with methods that can be called via RPC.
// a single server may have more than one Service.
type Service struct {
	Name    string
	Rcvr    reflect.Value
	Typ     reflect.Type
	Methods map[string]reflect.Method
}

func MakeService(rcvr interface{}) *Service {
	svc := &Service{}
	svc.Typ = reflect.TypeOf(rcvr)
	svc.Rcvr = reflect.ValueOf(rcvr)
	svc.Name = reflect.Indirect(svc.Rcvr).Type().Name()
	svc.Methods = map[string]reflect.Method{}

	for m := 0; m < svc.Typ.NumMethod(); m++ {
		method := svc.Typ.Method(m)
		mtype := method.Type
		mname := method.Name

		//fmt.Printf("%v pp %v ni %v 1k %v 2k %v no %v\n",
		//	mname, method.PkgPath, mtype.NumIn(), mtype.In(1).Kind(), mtype.In(2).Kind(), mtype.NumOut())

		if method.PkgPath != "" || // capitalized?
			mtype.NumIn() != 3 ||
			//mtype.In(1).Kind() != reflect.Ptr ||
			mtype.In(2).Kind() != reflect.Ptr ||
			mtype.NumOut() != 0 {
			// the method is not suitable for a handler
			//fmt.Printf("bad method: %v\n", mname)
		} else {
			// the method looks like a handler
			svc.Methods[mname] = method
		}
	}

	return svc
}

func (svc *Service) Dispatch(methname string, req protocol.ReqMsg) protocol.ReplyMsg {
	if method, ok := svc.Methods[methname]; ok {
		// prepare space into which to read the argument.
		// the Value's type will be a pointer to req.argsType.
		args := reflect.New(req.ArgsType)

		// decode the argument.
		ab := bytes.NewBuffer(req.Args)
		ad := sgob.NewDecoder(ab)
		ad.Decode(args.Interface())

		// allocate space for the reply.
		replyType := method.Type.In(2)
		replyType = replyType.Elem()
		replyv := reflect.New(replyType)

		// call the method.
		function := method.Func
		function.Call([]reflect.Value{svc.Rcvr, args.Elem(), replyv})

		// encode the reply.
		rb := new(bytes.Buffer)
		re := sgob.NewEncoder(rb)
		re.EncodeValue(replyv)

		return protocol.ReplyMsg{Ok: true, Reply: rb.Bytes()}
	} else {
		choices := []string{}
		for k := range svc.Methods {
			choices = append(choices, k)
		}
		log.Fatalf("labrpc.Service.dispatch(): unknown method %v in %v; expecting one of %v\n",
			methname, req.SvcMeth, choices)
		return protocol.ReplyMsg{Ok: false, Reply: nil}
	}
}