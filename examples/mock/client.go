package mock

import (
	"EIM"
	"EIM/logger"
	"EIM/tcp"
	"EIM/websocket"
	"context"
	"net"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

// ClientDemo 实现了Client接口
type ClientDemo struct{}

// Start demo入口方法
func (c *ClientDemo) Start(userID, protocol, addr string) {
	var cli EIM.Client

	// 1. 初始化客户端
	if protocol == "ws" {
		cli = websocket.NewClient(userID, "client", websocket.ClientOptions{})
		cli.SetDialer(&WebsocketDialer{})
	} else if protocol == "tcp" {
		cli = tcp.NewClient("test1", "client", tcp.ClientOptions{})
		cli.SetDialer(&TCPDialer{})
	}

	// 2. 建立连接
	err := cli.Connect(addr)
	if err != nil {
		logger.Error(err)
	}

	// 3. 发送消息然后退出
	count := 5
	go func() {
		for i := 0; i < count; i++ {
			message := []byte("hello")
			err := cli.Send(message)
			if err != nil {
				logger.Error(err)
				return
			}
			logger.Printf("%s send message [%s]", cli.ServiceID(), message)
			time.Sleep(time.Second)
		}
	}()

	// 4. 接收消息
	recv := 0
	for {
		frame, err := cli.Read()
		if err != nil {
			logger.Info(err)
			break
		}
		// 检查frame的OpCode
		if frame.GetOpCode() != EIM.OpBinary {
			continue
		}
		recv++
		logger.Warnf("%s receive message [%s]", cli.ServiceID(), frame.GetPayload())
		if recv == count {
			break
		}
	}
	cli.Close()
}

// ClientHandler 连接处理器, 实现MessageListener和StateListener
type ClientHandler struct{}

func (c *ClientHandler) Receive(ag EIM.Agent, payload []byte) {
	logger.Warnf("%s receive message [%s]", ag.ID(), string(payload))
}

func (c *ClientHandler) Disconnect(id string) error {
	logger.Warnf("Disconnect %s", id)
	return nil
}

// websocket拨号逻辑

// WebsocketDialer 实现Dialer接口, websocket拨号器
type WebsocketDialer struct {
}

func (d *WebsocketDialer) DialAndHandshake(ctx EIM.DialerContext) (net.Conn, error) {
	// 调用ws.Dial拨号
	conn, _, _, err := ws.Dial(context.TODO(), ctx.Address)
	if err != nil {
		return nil, err
	}
	// 发送用户认证消息, 实例是userId
	err = wsutil.WriteClientBinary(conn, []byte(ctx.Id))
	if err != nil {
		return nil, err
	}
	// 返回conn
	return conn, nil
}

// TCP拨号逻辑

// TCPDialer 实现Dialer接口, TCP拨号器
type TCPDialer struct {
}

func (d *TCPDialer) DialAndHandshake(ctx EIM.DialerContext) (net.Conn, error) {
	logger.Info("start dial: ", ctx.Address)
	// 调用net.Dial拨号
	conn, err := net.DialTimeout("tcp", ctx.Address, ctx.Timeout)
	if err != nil {
		return nil, err
	}
	// 发送用户认证消息, 实例是userId
	err = tcp.WriteFrame(conn, EIM.OpBinary, []byte(ctx.Id))
	if err != nil {
		return nil, err
	}
	// 返回conn
	return conn, nil
}
