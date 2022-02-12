package gobrpc

import (
	"io"
	"log"
	"bufio"
	"encoding/gob"
	"srpc/common/protocol"
)

type GobServerCodec struct {
	rwc    io.ReadWriteCloser
	dec    *gob.Decoder
	enc    *gob.Encoder
	encBuf *bufio.Writer
	closed bool
}

func (c *GobServerCodec) ReadRequestHeader(r *protocol.ReqMsg) error {
	return c.dec.Decode(r)
}

func (c *GobServerCodec) ReadRequestBody(body interface{}) error {
	return c.dec.Decode(body)
}

func (c *GobServerCodec) WriteResponse(r *protocol.ReplyMsg, body interface{}) (err error) {
	if err = c.enc.Encode(r); err != nil {
		if c.encBuf.Flush() == nil {
			log.Fatalf("rpc: gob error encoding response: %v", err)
			c.Close()
		}
		return
	}
	if err = c.enc.Encode(body); err != nil {
		if c.encBuf.Flush() == nil {
			log.Fatalf("rpc: gob error encoding body: %v", err)
			c.Close()
		}
		return
	}
	return c.encBuf.Flush()
}

func (c *GobServerCodec) Close() error {
	if c.closed {
		return nil
	}
	c.closed = true
	return c.rwc.Close()
}