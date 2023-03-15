package serv

import (
	"EIM"
	"EIM/container"
	"EIM/logger"
	"EIM/wire"
	"EIM/wire/pkt"
	"bytes"
	"strings"
	"time"

	"google.golang.org/protobuf/proto"
)

var log = logger.WithFields(logger.Fields{
	"module": wire.SNChat,
	"pkg":    "serv",
})

type ServHandler struct {
	r          *EIM.Router
	cache      EIM.SessionStorage
	dispatcher *ServerDispatcher
}

func NewServHandler(r *EIM.Router, cache EIM.SessionStorage) *ServHandler {
	return &ServHandler{
		r:          r,
		cache:      cache,
		dispatcher: &ServerDispatcher{},
	}
}

// Accept 节点处理链路, 用于握手处理, 接收新的连接
func (h *ServHandler) Accept(conn EIM.Conn, timeout time.Duration) (string, error) {
	logger.Infoln("enter")

	_ = conn.SetReadDeadline(time.Now().Add(timeout))
	frame, err := conn.ReadFrame()
	if err != nil {
		return "", err
	}
	var req pkt.InnerHandshakeReq
	_ = proto.Unmarshal(frame.GetPayload(), &req)

	log.Info("Accept -- ", req.ServiceId)

	// 把req.ServiceId当作ChannelId返回
	return req.ServiceId, nil
}

// Receive 接收网关发送过来的消息
func (h *ServHandler) Receive(ag EIM.Agent, payload []byte) {
	buf := bytes.NewBuffer(payload)
	packet, err := pkt.MustReadLogicPkt(buf)
	if err != nil {
		return
	}
	var session *pkt.Session
	if packet.Command == wire.CommandLoginSignIn {
		server, _ := packet.GetMeta(wire.MetaDestServer)
		session = &pkt.Session{
			ChannelId: packet.ChannelId,
			GateId:    server.(string),
			Tags:      []string{"AutoGenerated"},
		}
	} else {
		session, err = h.cache.Get(packet.ChannelId)
		if err == EIM.ErrSessionNil {
			_ = RespErr(ag, packet, pkt.Status_SessionNotFound)
			return
		} else if err != nil {
			_ = RespErr(ag, packet, pkt.Status_SystemException)
			return
		}
	}
	logger.Debugf("recv a message from %s  %s", session, &packet.Header)
	err = h.r.Serve(packet, h.dispatcher, h.cache, session)
	if err != nil {
		log.Warn(err)
	}
}

// RespErr 将错误信息推送会客户端
func RespErr(ag EIM.Agent, p *pkt.LogicPkt, status pkt.Status) error {
	packet := pkt.NewForm(&p.Header)
	packet.Status = status
	packet.Flag = pkt.Flag_Response

	return ag.Push(pkt.Marshal(packet))
}

// Disconnect 断开对应id的channel连接
func (h *ServHandler) Disconnect(id string) error {
	logger.Warnf("close event of %s", id)
	return nil
}

// ServerDispatcher 调度器
type ServerDispatcher struct{}

// Push 推送消息
func (s *ServerDispatcher) Push(gateway string, channels []string, pkt *pkt.LogicPkt) error {
	pkt.AddStringMeta(wire.MetaDestChannels, strings.Join(channels, ","))
	return container.Push(gateway, pkt)
}