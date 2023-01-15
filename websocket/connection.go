package websocket

import (
	"EIM"
	"github.com/gobwas/ws"
	"net"
)

// Frame 对ws.Frame的二次包装并实现EIM.Frame
type Frame struct {
	raw ws.Frame
}

func (f *Frame) SetOpCode(code EIM.OpCode) {
	f.raw.Header.OpCode = ws.OpCode(code)
}

func (f *Frame) GetOpCode() EIM.OpCode {
	return EIM.OpCode(f.raw.Header.OpCode)
}

func (f *Frame) SetPayload(payload []byte) {
	f.raw.Payload = payload
}

func (f *Frame) GetPayload() []byte {
	if f.raw.Header.Masked {
		ws.Cipher(f.raw.Payload, f.raw.Header.Mask, 0)
	}
	f.raw.Header.Masked = false
	return f.raw.Payload
}

// WsConn 对net.Conn二次包装并实现了EIM.Conn
type WsConn struct {
	net.Conn
}

func NewConn(conn net.Conn) *WsConn {
	return &WsConn{
		Conn: conn,
	}
}

func (c *WsConn) ReadFrame() (EIM.Frame, error) {
	f, err := ws.ReadFrame(c.Conn)
	if err != nil {
		return nil, err
	}
	return &Frame{raw: f}, nil
}

func (c *WsConn) WriteFrame(code EIM.OpCode, payload []byte) error {
	f := ws.NewFrame(ws.OpCode(code), true, payload)
	return ws.WriteFrame(c.Conn, f)
}

func (c *WsConn) Flush() error {
	return nil
}
