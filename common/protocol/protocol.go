package protocol

import "reflect"

// 消息格式

type ReqMsg struct {
	Endname  interface{} // name of sending ClientEnd
	SvcMeth  string      // e.g. "Raft.AppendEntries"
	ArgsType reflect.Type
	Args     []byte
	ReplyCh  chan ReplyMsg
}

type ReplyMsg struct {
	Ok    bool
	Reply []byte
}


// service配置文件格式
