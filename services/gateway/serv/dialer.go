package serv

import (
	"EIM"
	"EIM/logger"
	"EIM/tcp"
	"EIM/wire/pkt"
	"net"

	"google.golang.org/protobuf/proto"
)

// TcpDialer 拨号器, 用于服务之间的连接建立
type TcpDialer struct {
	ServiceId string
}

// DialAndHandshake 拨号握手
func (d *TcpDialer) DialAndHandshake(ctx EIM.DialerContext) (net.Conn, error) {
	// 拨号连接
	conn, err := net.DialTimeout("tcp", ctx.Address, ctx.Timeout)
	if err != nil {
		return nil, err
	}
	req := &pkt.InnerHandshakeReq{
		ServiceId: d.ServiceId,
	}
	logger.Infof("send req %v", req)
	// 将ServiceId发给对方
	bts, _ := proto.Marshal(req)
	err = tcp.WriteFrame(conn, EIM.OpBinary, bts)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// NewDialer 返回一个新Dialer
func NewDialer(serviceId string) EIM.Dialer {
	return &TcpDialer{
		ServiceId: serviceId,
	}
}
