package jsonrpc

import (
	"encoding/json"
	"srpc/common/protocol"
)

type JSONClientCodec struct {
	dec *json.Decoder
	enc *json.Encoder
}

func (c *JSONClientCodec) WriteRequest(r *protocol.ReqMsg, body interface{}) (err error) {
	if err = c.enc.Encode(r); err != nil {
		return
	}
	if err = c.enc.Encode(body); err != nil {
		return
	}
	return nil
}

func (c *JSONClientCodec) ReadResponseHeader(r *protocol.ReplyMsg) (err error) {
	return c.dec.Decode(r)
}

func (c *JSONClientCodec) ReadResponseBody(body interface{}) error {
	return c.dec.Decode(body)
}

func (c *JSONClientCodec) Close() error {
	return nil
}