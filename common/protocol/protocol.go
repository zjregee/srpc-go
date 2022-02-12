package protocol

import (
	"reflect"
)

type ReqMsg struct {
	Endname  interface{}
	SvcMeth  string
	ArgsType reflect.Type
	Args     []byte
	ReplyCh  chan ReplyMsg
}

type ReplyMsg struct {
	Ok    bool
	Reply []byte
}

type ClientCodec interface {
	WriteRequest(*ReplyMsg, interface{}) error
	ReadResponseHeader(*ReplyMsg) error
	ReadResponseBody(interface{}) error
	Close() error
}

type ServerCodec interface {
	ReadRequestHeader(*ReqMsg) error
	ReadRequestBody(interface{}) error
	WriteResponse(*ReplyMsg, interface{}) error
	Close() error
}