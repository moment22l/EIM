package EIM

import (
	"EIM/logger"
	"EIM/wire/pkt"
	"google.golang.org/protobuf/proto"
	"sync"
)

// Session 会话
type Session interface {
	GetChannelId() string
	GetGateId() string
	GetAccount() string
	GetRemoteIP() string
	GetApp() string
	GetTags() []string
}

// Context 上下文
type Context interface {
	Dispatcher
	SessionStorage
	Header() *pkt.Header
	Session() Session
	ReadBody(val proto.Message) error
	Resp(status pkt.Status, body proto.Message) error // 给消息发送方返回一条消息
	RespWithError(status pkt.Status, err error) error
	Dispatch(body proto.Message, recvs ...*Location) error
}

type HandlerFun func(ctx Context)

type HandlerChain []HandlerFun

// ContextImpl 上下文的具体实现
type ContextImpl struct {
	sync.Mutex
	Dispatcher
	SessionStorage

	handlers HandlerChain
	index    int
	request  *pkt.LoginPkt
	session  Session
}

// BuildContext 返回一个ContextImp
func BuildContext() Context {
	return &ContextImpl{}
}

// Next 调用下一个handler
func (c *ContextImpl) Next() {
	if c.index >= len(c.handlers) {
		return
	}
	f := c.handlers[c.index]
	c.index++
	if f == nil {
		logger.Warn("arrived unknown HandlerFunc")
		return
	}
	f(c)
}

func (c *ContextImpl) Header() *pkt.Header {
	//TODO implement me
	panic("implement me")
}

func (c *ContextImpl) Session() Session {
	//TODO implement me
	panic("implement me")
}

func (c *ContextImpl) ReadBody(val proto.Message) error {
	//TODO implement me
	panic("implement me")
}

// Resp 用于给消息发送方返回一条消息
func (c *ContextImpl) Resp(status pkt.Status, body proto.Message) error {
	packet := pkt.NewForm(&c.request.Header)
	packet.Status = status
	packet.WriteBody(body)
	packet.Flag = pkt.Flag_Response
	logger.Debugf("<-- Resp to %s command:%s  status: %v body: %s",
		c.Session().GetAccount(), &c.request.Header, status, body)
	err := c.Push(c.Session().GetGateId(), []string{c.Session().GetChannelId()}, packet)
	if err != nil {
		logger.Error(err)
	}
	return err
}

func (c *ContextImpl) RespWithError(status pkt.Status, err error) error {
	//TODO implement me
	panic("implement me")
}

// Dispatch 消息转发
func (c *ContextImpl) Dispatch(body proto.Message, recvs ...*Location) error {
	if len(recvs) == 0 {
		return nil
	}
	packet := pkt.NewForm(&c.request.Header)
	packet.Flag = pkt.Flag_Push
	packet.WriteBody(body)
	logger.Debugf("<-- Dispatch to %d users command:%s", len(recvs), &c.request.Header)

	group := make(map[string][]string)
	for _, recv := range recvs {
		if recv.ChannelId == c.request.GetChannelId() {
			continue
		}
		if _, ok := group[recv.GateId]; !ok {
			group[recv.GateId] = make([]string, 0)
		}
		group[recv.GateId] = append(group[recv.GateId], recv.ChannelId)
	}
	for gateway, ids := range group {
		err := c.Push(gateway, ids, packet)
		if err != nil {
			logger.Error(err)
		}
		return err
	}
	return nil
}

func (c *ContextImpl) reset() {
	c.handlers = nil
	c.index = 0
	c.request = nil
	c.session = nil
}
