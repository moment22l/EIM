package serv

import (
	"EIM"
	"EIM/container"
	"EIM/logger"
	"EIM/wire/pkt"
	"bytes"
	"google.golang.org/protobuf/proto"
	"time"
)

var log = logger.WithFields(logger.Fields{
	"service": "gateway",
	"pkt":     "serv",
})

type Handler struct {
	ServiceId string
	AddSecret string
}

// Accept 节点处理链路, 用于
func (h *Handler) Accept(conn EIM.Conn, timeout time.Duration) (string, error) {
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

// Receive 消息处理链路, 接受消息
func (h *Handler) Receive(ag EIM.Agent, payload []byte) {
	buf := bytes.NewBuffer(payload)
	packet, err := pkt.Read(buf)
	if err != nil {
		return
	}

	if basicPkt, ok := packet.(*pkt.BasicPkt); ok {
		if basicPkt.Code == pkt.CodePing {
			err = ag.Push(pkt.Marshal(&pkt.BasicPkt{Code: pkt.CodePong}))
			if err != nil {
				return
			}
		}
	}

	if loginPkt, ok := packet.(*pkt.LoginPkt); ok {
		loginPkt.ChannelId = ag.ID()
		err = container.Forward(loginPkt.ServiceName(), loginPkt)
		if err != nil {
			logger.WithFields(logger.Fields{
				"module": "handler",
				"id":     ag.ID(),
				"cmd":    loginPkt.Command,
				"dest":   loginPkt.Dest,
			}).Error(err)
		}
	}
}
