package tcp

import (
	"EIM"
	"EIM/logger"
	"context"
	"errors"
	"fmt"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

// ClientOptions 超时参数
type ClientOptions struct {
	Heartbeat time.Duration //登陆超时
	ReadWait  time.Duration //读超时
	WriteWait time.Duration //写超时
}

// Client tcp的Client实现
type Client struct {
	sync.Mutex
	EIM.Dialer
	id      string
	name    string
	conn    EIM.Conn
	state   int32
	once    sync.Once
	options ClientOptions
	Meta    map[string]string
}

// NewClient 创建新客户端
func NewClient(id, name string, opts ClientOptions) EIM.Client {
	return NewClientWithProps(id, name, make(map[string]string), opts)
}

func NewClientWithProps(id, name string, meta map[string]string, opts ClientOptions) EIM.Client {
	if opts.WriteWait == 0 {
		opts.WriteWait = EIM.DefaultWriteWait
	}
	if opts.ReadWait == 0 {
		opts.ReadWait = EIM.DefaultReadWait
	}
	cli := &Client{
		id:      id,
		name:    name,
		options: opts,
		Meta:    meta,
	}
	return cli
}

// ID 返回id
func (c *Client) ID() string {
	return c.id
}

// Name 返回name
func (c *Client) Name() string {
	return c.name
}

// Connect 客户端核心逻辑部分, 将客户端连接到对应服务端
func (c *Client) Connect(addr string) error {
	// 解析地址
	_, err := url.Parse(addr)
	if err != nil {
		return err
	}
	// 检查客户端是否已连接
	// 这里是一个CAS原子操作, 对比并设置值, 是并发安全的
	if !atomic.CompareAndSwapInt32(&c.state, 0, 1) {
		return fmt.Errorf("client has connected")
	}
	_, cancel := context.WithTimeout(context.TODO(), time.Second*10)
	defer cancel()
	rawconn, err := c.Dialer.DialAndHandshake(EIM.DialerContext{
		Id:      c.id,
		Name:    c.name,
		Address: addr,
		Timeout: EIM.DefaultLoginWait,
	})
	if err != nil {
		atomic.CompareAndSwapInt32(&c.state, 1, 0)
		return err
	}
	if rawconn == nil {
		return fmt.Errorf("conn is nil")
	}
	c.conn = NewConn(rawconn)

	if c.options.Heartbeat > 0 {
		go func() {
			err := c.heartbeatloop()
			if err != nil {
				logger.WithField("module", "tcp.client").Warn("heartbeatloop stopped - ", err)
			}
		}()
	}
	return nil
}

// SetDialer 设置握手逻辑
func (c *Client) SetDialer(dialer EIM.Dialer) {
	c.Dialer = dialer
}

// Send 发送消息到连接
func (c *Client) Send(payload []byte) error {
	if atomic.LoadInt32(&c.state) == 0 {
		return fmt.Errorf("connection is nil")
	}

	err := c.conn.SetWriteDeadline(time.Now().Add(c.options.WriteWait))
	if err != nil {
		return err
	}
	return c.conn.WriteFrame(EIM.OpBinary, payload)
}

// Read 从连接中读取消息
func (c *Client) Read() (EIM.Frame, error) {
	if c.conn == nil {
		return nil, errors.New("connection is nil")
	}
	if c.options.ReadWait > 0 {
		_ = c.conn.SetReadDeadline(time.Now().Add(c.options.ReadWait))
	}
	frame, err := c.conn.ReadFrame()
	if err != nil {
		return nil, err
	}
	if frame.GetOpCode() == EIM.OpClose {
		return nil, errors.New("remote side close the channel")
	}
	return frame, nil
}

// heartbeatloop 心跳逻辑
func (c *Client) heartbeatloop() error {
	tick := time.NewTicker(c.options.Heartbeat)
	for range tick.C {
		err := c.ping()
		if err != nil {
			return err
		}
	}
	return nil
}

// ping ping一个心跳给服务端
func (c *Client) ping() error {
	logger.WithField("module", "tcp.client").Tracef("%s send ping to server", c.id)

	err := c.conn.SetWriteDeadline(time.Now().Add(c.options.WriteWait))
	if err != nil {
		return err
	}

	return c.conn.WriteFrame(EIM.OpPing, nil)
}

// Close 关闭客户端
func (c *Client) Close() {
	c.once.Do(func() {
		if c.conn == nil {
			return
		}
		// 优雅退出
		_ = WriteFrame(c.conn, EIM.OpClose, nil)
		c.conn.Close()
		atomic.CompareAndSwapInt32(&c.state, 1, 0)
	})
}
