package client

import (
	"log"
	"net"
	"fmt"
	"sync"
	"bytes"
	"strings"
	"reflect"
	"srpc/common/protocol"
	"srpc/common/sgob"
)

type ClientEnd struct {
	mu      sync.Mutex
	endname interface{}
	ch      chan protocol.ReqMsg
	done    chan struct{}
	network *Network
	config  *Config
}

type Network struct {
	Registry_enabled bool
	Registry_ip      string
	Registry_port    string
	Services         []*Service
}

type Service struct {
	Service_name    string
	Method_name     string
	Server_ip       string
	Server_port     string
	Service_enabled bool
}

func MakeClientEnd() (*ClientEnd, error) {
	e := &ClientEnd{}
	return e, nil
}

func MakeClientEndFromConfig(fName string) (*ClientEnd, error) {
	var err error
	e := &ClientEnd{}
	e.config, err = NewConfig(fName, &JSONConfigFormat{})
	if err != nil {
		return nil, err
	}
	e.network = e.config.Format.TransferToNetWork()
	return e, nil
}

func MakeClientEndFromConfigText(text string) (*ClientEnd, error) {
	var err error
	e := &ClientEnd{}
	e.config, err = NewConfigFromText(text, &JSONConfigFormat{})
	if err != nil {
		return nil, err
	}
	e.network = e.config.Format.TransferToNetWork()
	return e, nil
} 

func (e *ClientEnd) RefrshConfig(fName string) error {
	var err error
	e.config, err = NewConfig(fName, &JSONConfigFormat{})
	if err != nil {
		return err
	}
	e.network = e.config.Format.TransferToNetWork()
	return nil
}

func (e *ClientEnd) RefreshConfigFromText(text string) error {
	var err error
	e.config, err = NewConfigFromText(text, &JSONConfigFormat{})
	if err != nil {
		return err
	}
	e.network = e.config.Format.TransferToNetWork()
	return nil
}

func (e *ClientEnd) SetRegistry(ip, port string) error {
	e.network.Registry_ip = ip
	e.network.Registry_port = port
	e.network.Registry_enabled = true
	return nil
}

func (e *ClientEnd) Call(svcMeth string, args interface{}, reply interface{}) bool {
	service := e.chooseService(svcMeth)
	if service == nil {
		e.pullService(svcMeth)
		return false
	}

	e.send([]byte{}, service.Server_ip, service.Server_port)

	req := protocol.ReqMsg{}
	req.Endname = e.endname
	req.SvcMeth = svcMeth
	req.ArgsType = reflect.TypeOf(args)
	req.ReplyCh = make(chan protocol.ReplyMsg)

	qb := new(bytes.Buffer)
	qe := sgob.NewEncoder(qb)
	if err := qe.Encode(args); err != nil {
		panic(err)
	}
	req.Args = qb.Bytes()

	select {
	case e.ch <- req:
	case <-e.done:
		return false
	}

	rep := <-req.ReplyCh
	if rep.Ok {
		rb := bytes.NewBuffer(rep.Reply)
		rd := sgob.NewDecoder(rb)
		if err := rd.Decode(reply); err != nil {
			log.Fatalf("ClientEnd.Call(): decode reply: %v\n", err)
		}
		return true
	} else {
		return false
	}
}

func (e *ClientEnd) Close() {
	e.done <- struct{}{}
}

func (e *ClientEnd) send(data []byte, ip, port string) ([]byte, error) {
	address := fmt.Sprintf("%s:%s", ip, port)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Fatalf("dial failed, err: %v", err)
		return nil, err
	}
	defer conn.Close()
	_, err = conn.Write(data)
	if err != nil {
		return nil, err
	}
	buf := [512]byte{}
	n, err := conn.Read(buf[:])
	if err != nil {
		log.Fatalf("recv failed, err: %v", err)
		return nil, err
	}
	return buf[:n], nil
}

func (e *ClientEnd) chooseService(svcMeth string) *Service {
	params := strings.Split(svcMeth, ".")
	if len(params) != 2 {
		return nil
	}
	serviceName := params[0]
	methodName := params[1]
	services := []*Service{}
	for _, service := range e.network.Services {
		if service.Service_name == serviceName && service.Method_name == methodName {
			services = append(services, service)
		}
	}
	if len(services) == 0 {
		return nil
	}
	return e.scheduleService(services)
}

func (e *ClientEnd) scheduleService(servics []*Service) *Service {
	return servics[0]
}

func (e *ClientEnd) pullServices() {

}

func (e *ClientEnd) pullService(svcMeth string) {

}