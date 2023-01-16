package tcp

import (
	"EIM"
	"EIM/wire/endian"
	"io"
	"net"
)

// Frame Tcp帧并实现EIM.Frame
type Frame struct {
	OpCode  EIM.OpCode
	Payload []byte
}

// SetOpCode 设置OpCode
func (f *Frame) SetOpCode(code EIM.OpCode) {
	f.OpCode = code
}

// GetOpCode 获取OpCode
func (f *Frame) GetOpCode() EIM.OpCode {
	return f.OpCode
}

// SetPayload 设置Payload
func (f *Frame) SetPayload(payload []byte) {
	f.Payload = payload
}

// GetPayload 获取Payload
func (f *Frame) GetPayload() []byte {
	return f.Payload
}

// TcpConn Tcp连接, 二次包装net.Conn并实现EIM.Conn
type TcpConn struct {
	net.Conn
}

// NewConn 创建一个新的TcpConn
func NewConn(conn net.Conn) *TcpConn {
	return &TcpConn{
		Conn: conn,
	}
}

// ReadFrame 从c.Conn中读取一帧
func (c *TcpConn) ReadFrame() (EIM.Frame, error) {
	opcode, err := endian.ReadUint8(c.Conn)
	if err != nil {
		return nil, err
	}
	payload, err := endian.ReadBytes(c.Conn)
	if err != nil {
		return nil, err
	}
	return &Frame{
		OpCode:  EIM.OpCode(opcode),
		Payload: payload,
	}, nil
}

// WriteFrame 向c.Conn中写入一帧
func (c *TcpConn) WriteFrame(code EIM.OpCode, payload []byte) error {
	return WriteFrame(c.Conn, code, payload)
}

// Flush 刷新
func (c *TcpConn) Flush() error {
	return nil
}

// WriteFrame 往w中写入一个frame
func WriteFrame(w io.Writer, code EIM.OpCode, payload []byte) error {
	if err := endian.WriteUint8(w, uint8(code)); err != nil {
		return err
	}
	if err := endian.WriteBytes(w, payload); err != nil {
		return err
	}
	return nil
}
