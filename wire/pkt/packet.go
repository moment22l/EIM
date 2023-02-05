package pkt

import (
	"EIM/wire/endian"
	"google.golang.org/protobuf/proto"
	"io"
)

// LoginPkt 逻辑协议消息包(网关对外的client消息结构)
type LoginPkt struct {
	Header
	Body []byte `json:"body,omitempty"`
}

// Decode 从r中读取若干字节到LoginPkt并解包
func (p *LoginPkt) Decode(r io.Reader) error {
	// 读取Header
	headerBytes, err := endian.ReadBytes(r)
	if err != nil {
		return err
	}
	if err = proto.Unmarshal(headerBytes, &p.Header); err != nil {
		return err
	}
	// 读取Body
	p.Body, err = endian.ReadBytes(r)
	if err != nil {
		return err
	}
	return nil
}

// Encode 封包并写入w
func (p *LoginPkt) Encode(w io.Writer) error {
	headerBytes, err := proto.Marshal(&p.Header)
	if err != nil {
		return err
	}
	if err = endian.WriteBytes(w, headerBytes); err != nil {
		return err
	}
	if err = endian.WriteBytes(w, p.Body); err != nil {
		return err
	}
	return nil
}
