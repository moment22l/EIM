package EIM

import (
	"EIM/naming"
	"context"
	"net"
	"time"
)

const (
	DefaultReadWait  = time.Minute * 3
	DefaultWriteWait = time.Second * 10
	DefaultLoginWait = time.Second * 10
	DefaultHeartbeat = time.Second * 55
)

// Server 接口
type Server interface {
	naming.ServiceRegistration          // 服务
	SetAcceptor(Acceptor)               // 用于设置一个Acceptor
	SetMessageListener(MessageListener) // 用于设置一个MessageListener(上行消息监听器)
	SetStateListener(StateListener)     // 用于设置一个StateListener(连接状态监听服务)
	SetReadWait(time.Duration)          // 用于设置一个连接读超时等待时间
	SetChannelMap(ChannelMap)           // 用于设置一个ChannelMap(连接管理器)

	// Start 用于在内部实现网络端口的监听和接收连接，
	// 并完成一个Channel的初始化过程。
	Start() error
	// Push 消息到指定的Channel中
	// string channelID
	// []byte 序列化之后的消息数据
	Push(string, []byte) error
	Shutdown(context.Context) error // 服务下线，关闭连接
}

// Acceptor 调用Accept方法, 让上层业务处理握手相关工作
type Acceptor interface {
	Accept(Conn, time.Duration) (string, error) // 返回的是一个ChannelID(唯一通道标识)
}

// MessageListener 消息监听器
type MessageListener interface {
	Receive(Agent, []byte)
}

// Agent 发送方
type Agent interface {
	ID() string
	Push([]byte) error
}

// StateListener 状态监听器
type StateListener interface {
	Disconnect(string) error
}

// Channel 客户端接口
type Channel interface {
	Conn
	Agent
	Close() error // 重写net.Conn中的Close方法
	Readloop(lst MessageListener) error
	SetWriteWait(time.Duration)
	SetReadWait(time.Duration)
}

// Frame 数据包
type Frame interface {
	SetOpCode(OpCode)
	GetOpCode() OpCode
	SetPayload([]byte)
	GetPayload() []byte
}

// Conn connection 对net.Conn进行二次包装, 将读写操作封装进连接中
type Conn interface {
	net.Conn
	ReadFrame() (Frame, error)
	WriteFrame(OpCode, []byte) error
	Flush() error
}

// OpCode 帧类型
type OpCode int

const (
	OpContinuation OpCode = 0x0
	OpText         OpCode = 0x1
	OpBinary       OpCode = 0x2
	OpClose        OpCode = 0x8
	OpPing         OpCode = 0x9
	OpPong         OpCode = 0xa
)

// Client 客户端接口
type Client interface {
	ID() string
	Name() string
	Connect(string) error
	SetDialer(Dialer)
	Send([]byte) error
	Read() (Frame, error) // 底层复用了kim.Conn, 所以直接返回Frame
	Close()
}

// Dialer 拨号器, 在Connect中被调用, 完成连接的建立和握手
type Dialer interface {
	DialAndHandshake(DialerContext) (net.Conn, error)
}

// DialerContext 拨号和握手所需信息
type DialerContext struct {
	Id      string
	Name    string
	Address string
	Timeout time.Duration
}
