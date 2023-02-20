package pkt

import (
	"EIM/wire"
	"bytes"
	"errors"
	"fmt"
	"io"
	"reflect"
)

// Packet 协议消息接口
type Packet interface {
	Decode(r io.Reader) error
	Encode(w io.Writer) error
}

// Read 利用魔数区分协议, 解包
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

// MustReadLoginPkt 必须读取一个LoginPkt, 否则返回错误信息
func MustReadLoginPkt(r io.Reader) (*LoginPkt, error) {
	val, err := Read(r)
	if err != nil {
		return nil, err
	}
	if lp, ok := val.(*LoginPkt); ok {
		return lp, nil
	}
	return nil, fmt.Errorf("this packet is not a Login packet")
}

// MustReadBasicPkt 必须读取一个BasicPkt, 否则返回错误信息
func MustReadBasicPkt(r io.Reader) (*BasicPkt, error) {
	val, err := Read(r)
	if err != nil {
		return nil, err
	}
	if lp, ok := val.(*BasicPkt); ok {
		return lp, nil
	}
	return nil, fmt.Errorf("this packet is not a Login packet")
}

// Marshal 把魔数Magic封装到消息的头部, 封包(利用到了golang的反射机制)
func Marshal(p Packet) []byte {
	buf := new(bytes.Buffer)
	kind := reflect.TypeOf(p).Elem()

	if kind.AssignableTo(reflect.TypeOf(LoginPkt{})) {
		_, _ = buf.Write(wire.MagicLoginPkt[:])
	} else if kind.AssignableTo(reflect.TypeOf(BasicPkt{})) {
		_, _ = buf.Write(wire.MagicBasicPkt[:])
	}
	_ = p.Encode(buf)
	return buf.Bytes()
}
