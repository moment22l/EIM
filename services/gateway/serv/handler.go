package serv

import (
	"EIM"
	"EIM/container"
	"EIM/logger"
	"EIM/wire"
	"EIM/wire/pkt"
	"EIM/wire/token"
	"bytes"
	"fmt"
	"regexp"
	"time"
)

var log = logger.WithFields(logger.Fields{
	"service": "gateway",
	"pkt":     "serv",
})

type Handler struct {
	ServiceID string
}

// Accept 节点处理链路, 用于握手处理
func (h *Handler) Accept(conn EIM.Conn, timeout time.Duration) (string, error) {
	log := logger.WithFields(logger.Fields{
		"ServiceID": h.ServiceID,
		"module":    "Handler",
		"handler":   "Accept",
	})
	log.Infoln("enter")
	// 读取登录包
	_ = conn.SetReadDeadline(time.Now().Add(timeout))
	frame, err := conn.ReadFrame()
	if err != nil {
		return "", err
	}
	buf := bytes.NewBuffer(frame.GetPayload())
	req, err := pkt.MustReadLoginPkt(buf)
	if err != nil {
		return "", err
	}
	// 检测数据包是否为登录包
	if req.Command != wire.CommandLoginSignIn {
		resp := pkt.NewForm(&req.Header)
		resp.Status = pkt.Status_InvalidCommand
		_ = conn.WriteFrame(EIM.OpBinary, pkt.Marshal(resp))
		return "", fmt.Errorf("must be a InvalidCommand command")
	}
	// 对body进行反序列化
	var login pkt.LoginReq
	err = req.ReadBody(&login)
	if err != nil {
		return "", err
	}
	// 使用DefaultSecret解析token
	tk, err := token.Parse(token.DefaultKey, login.Token)
	if err != nil {
		// token无效则返回给SDK一个Unauthorized消息
		resp := pkt.NewForm(&req.Header)
		resp.Status = pkt.Status_Unauthorized
		_ = conn.WriteFrame(EIM.OpBinary, pkt.Marshal(resp))
		return "", err
	}
	// 生成全局唯一channelID
	id := generateChannelID(h.ServiceID, tk.Account)
	// 填写req包相关信息
	req.WriteBody(&pkt.Session{
		ChannelId: id,
		GateId:    h.ServiceID,
		Account:   tk.Account,
		App:       tk.App,
		RemoteIP:  getIp(conn.RemoteAddr().String()),
	})
	// 把req转发给login服务
	err = container.Forward(wire.SNLogin, req)
	if err != nil {
		return "", err
	}
	return id, nil
}

// Receive 消息处理链路, 接受消息
func (h *Handler) Receive(ag EIM.Agent, payload []byte) {
	buf := bytes.NewBuffer(payload)
	packet, err := pkt.Read(buf)
	if err != nil {
		log.Error(err)
		return
	}

	if basicPkt, ok := packet.(*pkt.BasicPkt); ok {
		if basicPkt.Code == pkt.CodePing {
			err = ag.Push(pkt.Marshal(&pkt.BasicPkt{Code: pkt.CodePong}))
			if err != nil {
				return
			}
		}
		return
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

// Disconnect 断开对应id的channel
func (h *Handler) Disconnect(channelId string) error {
	log.Infof("disconnect %s", channelId)

	logout := pkt.New(wire.CommandLoginSignOut, pkt.WithChannelId(channelId))
	err := container.Push(wire.SNLogin, logout)
	if err != nil {
		logger.WithFields(logger.Fields{
			"module": "handler",
			"id":     channelId,
		}).Error(err)
	}
	return nil
}

var ipExp = regexp.MustCompile("\\:[0-9]+$")

func getIp(remoteAddr string) string {
	if remoteAddr == "" {
		return ""
	}
	return ipExp.ReplaceAllString(remoteAddr, "")
}

// generateChannelID 产生唯一的channelID
func generateChannelID(serviceID string, account string) string {
	return fmt.Sprintf("%s_%s_%d", serviceID, account, wire.Seq.Next())
}
