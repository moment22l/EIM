package dialer

import (
	"EIM"
	"EIM/logger"
	"EIM/wire"
	"EIM/wire/pkt"
	"EIM/wire/token"
	"bytes"
	"context"
	"fmt"
	"net"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

type ClientDialer struct{}

func (d *ClientDialer) DialAndHandshake(ctx EIM.DialerContext) (net.Conn, error) {
	logger.Info("DialAndHandShake called")
	// 拨号
	conn, _, _, err := ws.Dial(context.TODO(), ctx.Address)
	if err != nil {
		return nil, err
	}
	// 生成token
	tk, err := token.Generate(token.DefaultKey, &token.Token{
		Account: ctx.Id,
		App:     "EIM",
		Exp:     time.Now().AddDate(0, 0, 1).Unix(),
	})
	if err != nil {
		return nil, err
	}
	// 发出一条CommandLoginSignIn消息
	loginReq := pkt.New(wire.CommandLoginSignIn).WriteBody(&pkt.LoginReq{
		Token: tk,
	})
	err = wsutil.WriteClientBinary(conn, pkt.Marshal(loginReq))
	if err != nil {
		return nil, err
	}
	// 等待loginReq对应的Resp
	logger.Info("waiting for login response")
	_ = conn.SetReadDeadline(time.Now().Add(ctx.Timeout))
	frame, err := ws.ReadFrame(conn)
	if err != nil {
		return nil, err
	}
	ack, err := pkt.MustReadLogicPkt(bytes.NewBuffer(frame.Payload))
	if err != nil {
		return nil, err
	}
	// 判断是否登录成功
	if ack.Status != pkt.Status_Success {
		return nil, fmt.Errorf("login failed: %v", &ack.Header)
	}
	var resp = new(pkt.LoginResp)
	_ = ack.ReadBody(resp)

	logger.Info("login ", resp.GetChannelId())
	return conn, nil
}
