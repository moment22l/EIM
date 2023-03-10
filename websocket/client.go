package websocket

import (
	"EIM"
	"EIM/logger"
	"errors"
	"fmt"
	"net"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

// ClientOptions 超时参数
type ClientOptions struct {
	Heartbeat time.Duration //登录超时
	ReadWait  time.Duration //读超时
	WriteWait time.Duration //写超时
}

// Client websocket的Client实现
type Client struct {
	sync.Mutex
	EIM.Dialer
	once    sync.Once
	id      string
	name    string
	conn    net.Conn
	state   int32
	options ClientOptions
	Meta    map[string]string
}

// NewClient 创建一个新客户端
func NewClient(id, name string, options ClientOptions) EIM.Client {
	return NewClientWithProps(id, name, make(map[string]string), options)
}

// NewClientWithProps 新建一个带有meta的客户端
func NewClientWithProps(id, name string, meta map[string]string, options ClientOptions) EIM.Client {
	if options.ReadWait == 0 {
		options.ReadWait = EIM.DefaultReadWait
	}
	if options.WriteWait == 0 {
		options.WriteWait = EIM.DefaultWriteWait
	}
	cli := &Client{
		id:      id,
		name:    name,
		options: options,
		Meta:    meta,
	}
	return cli
}

// Connect 连接到服务端Server
func (c *Client) Connect(addr string) error {
	// 解析地址
	_, err := url.Parse(addr)
	if err != nil {
		return err
	}
	// 查看客户端是否已处于连接状态
	if !atomic.CompareAndSwapInt32(&c.state, 0, 1) {
		return fmt.Errorf("client has connected")
	}
	// 拨号和握手
	conn, err := c.Dialer.DialAndHandshake(EIM.DialerContext{
		Id:      c.id,
		Name:    c.name,
		Address: addr,
		Timeout: EIM.DefaultLoginWait,
	})
	if err != nil {
		atomic.CompareAndSwapInt32(&c.state, 1, 0)
		return err
	}
	if conn == nil {
		return fmt.Errorf("conn is nil")
	}
	c.conn = conn

	if c.options.Heartbeat > 0 {
		go func() {
			err = c.heartbeatloop(conn)
			if err != nil {
				logger.Error("heartbeatloop stopped ", err)
			}
		}()
	}
	return nil
}

// ServiceID 返回id
func (c *Client) ServiceID() string {
	return c.id
}

// ServiceName 返回name
func (c *Client) ServiceName() string {
	return c.name
}

// SetDialer 设置拨号器
func (c *Client) SetDialer(dialer EIM.Dialer) {
	c.Dialer = dialer
}

// GetMeta 获取meta
func (c *Client) GetMeta() map[string]string {
	return c.Meta
}

// Close 关闭连接
func (c *Client) Close() {
	c.once.Do(func() {
		if c.conn == nil {
			return
		}
		// 优雅关闭连接
		_ = wsutil.WriteClientMessage(c.conn, ws.OpClose, nil)

		err := c.conn.Close()
		if err != nil {
			return
		}
		atomic.CompareAndSwapInt32(&c.state, 1, 0)
	})
}

// heartbeatloop 启用一个定时器发送心跳包
func (c *Client) heartbeatloop(conn net.Conn) error {
	tick := time.NewTicker(c.options.Heartbeat)
	for range tick.C {
		// 每隔tick时间ping一个心跳包给服务器
		err := c.ping(conn)
		if err != nil {
			return err
		}
	}
	return nil
}

// ping 发送一个心跳包给服务器
func (c *Client) ping(conn net.Conn) error {
	c.Lock()
	defer c.Unlock()
	// 通过conn.SetWriteDeadline重置写超时, 使得如果连接异常在发送端就可以感知到
	err := conn.SetWriteDeadline(time.Now().Add(c.options.WriteWait))
	if err != nil {
		return err
	}
	return wsutil.WriteClientMessage(conn, ws.OpPing, nil)
}

// Send 发送消息到连接
func (c *Client) Send(payload []byte) error {
	if atomic.LoadInt32(&c.state) == 0 {
		return fmt.Errorf("connection is nil")
	}
	c.Lock()
	defer c.Unlock()
	err := c.conn.SetWriteDeadline(time.Now().Add(c.options.WriteWait))
	if err != nil {
		return err
	}
	// 客户端消息需要使用MASK
	return wsutil.WriteClientMessage(c.conn, ws.OpBinary, payload)
}

// Read 从连接中读取消息(这个方法不是线程安全的)
func (c *Client) Read() (EIM.Frame, error) {
	if c.conn == nil {
		return nil, errors.New("connection is nil")
	}
	if c.options.ReadWait > 0 {
		_ = c.conn.SetReadDeadline(time.Now().Add(c.options.ReadWait))
	}
	frame, err := ws.ReadFrame(c.conn)
	if err != nil {
		return nil, err
	}
	if frame.Header.OpCode == ws.OpClose {
		return nil, errors.New("remote side close the channel")
	}
	return &Frame{raw: frame}, nil
}
