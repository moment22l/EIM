package pkt

import (
	"EIM/wire"
	"errors"
	"io"
)

// Packet 协议消息接口
type Packet interface {
	Decode(r io.Reader) error
	Encode(w io.Writer) error
}

// Read 利用魔数区分协议并解包
func Read(r io.Reader) (interface{}, error) {
	magic := wire.Magic{}
	_, err := io.ReadFull(r, magic[:])
	if err != nil {
		return nil, err
	}
	switch magic {
	case wire.MagicLoginPkt:
		p := new(LoginPkt)
		if err := p.Decode(r); err != nil {
			return nil, err
		}
		return p, nil
	case wire.MagicBasicPkt:
		p := new(BasicPkt)
		if err := p.Decode(r); err != nil {
			return nil, err
		}
		return p, nil
	default:
		return nil, errors.New("magic code is incorrect")
	}
}

// Marshal 把魔数Magic封装到消息的头部(利用到了golang的反射机制)
func Marshal(p Packet) []byte {
	// TODO: implement me
	return nil
}
