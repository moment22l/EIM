package mock

import (
	"EIM"
	"EIM/logger"
	"EIM/naming"
	"EIM/tcp"
	"EIM/websocket"
	"errors"
	"time"
)

// ServerDemo 模拟实现Server接口
type ServerDemo struct{}

// Start demo入口方法
func (s *ServerDemo) Start(id, protocol, addr string) {
	var srv EIM.Server
	Service := &naming.DefaultService{
		Id:       id,
		Protocol: protocol,
	}
	if protocol == "ws" {
		srv = websocket.NewServer(addr, Service)
	} else if protocol == "tcp" {
		srv = tcp.NewServer(addr, Service)
	}

	handler := &ServerHandler{}

	srv.SetAcceptor(handler)
	srv.SetMessageListener(handler)
	srv.SetStateListener(handler)

	err := srv.Start()
	if err != nil {
		panic(err)
	}
}

// ServerHandler 连接处理器
type ServerHandler struct{}

// Accept 实现Acceptor接口, accept a connection
func (h *ServerHandler) Accept(conn EIM.Conn, timeout time.Duration) (string, error) {
	// 1. 读取客户端发来的鉴权数据报
	frame, err := conn.ReadFrame()
	if err != nil {
		return "", err
	}
	logger.Info("recv", frame.GetOpCode())
	// 2. 解析数据包内容
	userID := string(frame.GetPayload())
	// 3. 鉴权, 这里做了一个假的验证
	if userID == "" {
		return "", errors.New("user id is invalid")
	}
	return userID, nil
}

// Receive 实现MessageListener接口, receive default listener
func (h *ServerHandler) Receive(ag EIM.Agent, payload []byte) {
	// ack即为需要推送到客户端的消息
	ack := string(payload) + " from server "
	_ = ag.Push([]byte(ack))
}

// Disconnect 实现StateListener接口, disconnect default listener
func (h *ServerHandler) Disconnect(id string) error {
	logger.Warnf("disconnect %s", id)
	return nil
}
