package pkt

import (
	"EIM/wire/endian"
	"fmt"
	"io"
	"strconv"
	"strings"

	"google.golang.org/protobuf/proto"
)

// LogicPkt 逻辑协议消息包(网关对外的client消息结构)
type LogicPkt struct {
	Header
	Body []byte `json:"body,omitempty"`
}

// HeaderOption 一系列设置Header参数的函数
type HeaderOption func(*Header)

// WithStatus 设置状态码
func WithStatus(status Status) HeaderOption {
	return func(h *Header) {
		h.Status = status
	}
}

// WithSeq 设置序列号
func WithSeq(seq uint32) HeaderOption {
	return func(h *Header) {
		h.Sequence = seq
	}
}

// WithChannelId 设置连接标识
func WithChannelId(channelId string) HeaderOption {
	return func(h *Header) {
		h.ChannelId = channelId
	}
}

// WithDest 设置目标(群/用户)
func WithDest(dest string) HeaderOption {
	return func(h *Header) {
		h.Dest = dest
	}
}

// New 根据command和options, new一个空白的LogicPkt
func New(command string, options ...HeaderOption) *LogicPkt {
	pkt := &LogicPkt{}
	pkt.Command = command

	for _, option := range options {
		option(&pkt.Header)
	}
	return pkt
}

// NewForm 根据一个Header创建一个LogicPkt
func NewForm(h *Header) *LogicPkt {
	pkt := &LogicPkt{}
	pkt.Header = Header{
		Command:   h.Command,
		ChannelId: h.ChannelId,
		Sequence:  h.Sequence,
		Status:    h.Status,
		Dest:      h.Dest,
	}
	return pkt
}

// Decode 从r中读取若干字节到LogicPkt并解包
func (p *LogicPkt) Decode(r io.Reader) error {
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

// Encode 封包并将p写入w
func (p *LogicPkt) Encode(w io.Writer) error {
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

// ReadBody 读取p中的body
func (p *LogicPkt) ReadBody(val proto.Message) error {
	return proto.Unmarshal(p.Body, val)
}

// WriteBody 将val序列化后写入到p的Body中
func (p *LogicPkt) WriteBody(val proto.Message) *LogicPkt {
	if val == nil {
		return p
	}
	p.Body, _ = proto.Marshal(val)
	return p
}

// StringBody 返回string形式的body
func (p *LogicPkt) StringBody() string {
	return string(p.Body)
}

// String 返回string形式的p
func (p *LogicPkt) String() string {
	return fmt.Sprintf("header:%v body:%dbits", &p.Header, len(p.Body))
}

// ServiceName 返回x中的服务名称
func (x *Header) ServiceName() string {
	arr := strings.SplitN(x.Command, ".", 2)
	if len(arr) <= 1 {
		return "default"
	}
	return arr[0]
}

// AddMeta 向p中的Meta数组中添加元素
func (p *LogicPkt) AddMeta(meta ...*Meta) {
	p.Meta = append(p.Meta, meta...)
}

// AddStringMeta 向p中的Meta数组中添加string形式的元素
func (p *LogicPkt) AddStringMeta(key, value string) {
	p.AddMeta(&Meta{
		Key:   key,
		Value: value,
		Type:  MetaType_string,
	})
}

// FindMeta 查询对应的Meta
func FindMeta(meta []*Meta, key string) (interface{}, bool) {
	for _, m := range meta {
		if m.Key == key {
			switch m.Type {
			case MetaType_int:
				v, _ := strconv.Atoi(m.Value)
				return v, true
			case MetaType_float:
				v, _ := strconv.ParseFloat(m.Value, 64)
				return v, true
			}
			return m.Value, true
		}
	}
	return nil, false
}

// GetMeta 从p的Meta数组中找到key所对应的值
func (p *LogicPkt) GetMeta(key string) (interface{}, bool) {
	return FindMeta(p.Meta, key)
}

// DelMeta 删除p的Meta数组中key所对应的元素
func (p *LogicPkt) DelMeta(key string) {
	for i, m := range p.Meta {
		if m.Key == key {
			length := len(p.Meta)
			if i < length-1 {
				copy(p.Meta[i:], p.Meta[i+1:])
			}
			p.Meta = p.Meta[:length-1]
		}
	}
}
