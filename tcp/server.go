package tcp

import (
	"EIM"
	"context"
	"errors"
	"fmt"
	"github.com/segmentio/ksuid"
	"net"
	"sync"
	"time"
)

// ServerOptions 超时参数
type ServerOptions struct {
	loginwait time.Duration // 登录超时
	readwait  time.Duration // 读超时
	writewait time.Duration // 写超时
}

// Server tcp的Server实现
type Server struct {
	listen string
	naming.ServiceRegistration
	EIM.Acceptor
	EIM.MessageListener
	EIM.StateListener
	EIM.ChannelMap
	once    sync.Once
	options ServerOptions
}

// SetAcceptor 设置接收器Acceptor
func (s *Server) SetAcceptor(acceptor EIM.Acceptor) {
	s.Acceptor = acceptor
}

// SetMessageListener 设置信息监听器MessageListener
func (s *Server) SetMessageListener(listener EIM.MessageListener) {
	s.MessageListener = listener
}

// SetStateListener 设置装填监听器StateListener
func (s *Server) SetStateListener(listener EIM.StateListener) {
	s.StateListener = listener
}

// SetReadWait 设置读超时
func (s *Server) SetReadWait(readwait time.Duration) {
	s.options.readwait = readwait
}

// SetChannelMap 设置连接管理表
func (s *Server) SetChannelMap(channelMap EIM.ChannelMap) {
	s.ChannelMap = channelMap
}

// Start Server的核心逻辑部分, 开启服务器
func (s *Server) Start() error {
	log := logger.WithFields(logger.Fields{
		"module": "tcp.server",
		"listen": s.listen,
		"id":     s.ServiceID(),
	})

	if s.StateListener == nil {
		return fmt.Errorf("stateListener is nil")
	}
	if s.Acceptor == nil {
		s.Acceptor = new(defaultAcceptor)
	}
	// 1. 启用连接监听
	lst, err := net.Listen("tcp", s.listen)
	if err != nil {
		return err
	}
	log.Info("started")
	for {
		// 2. 接收新的连接
		rawconn, err := lst.Accept()
		if err != nil {
			rawconn.Close()
			log.Warn(err)
			continue
		}
		go func(rawconn net.Conn) {
			conn := NewConn(rawconn)
			// 3. 交给上层处理认证等逻辑
			id, err := s.Accept(conn, s.options.loginwait)
			if err != nil {
				_ = conn.WriteFrame(EIM.OpClose, []byte(err.Error()))
				conn.Close()
				return
			}
			if _, ok := s.Get(id); ok {
				log.Warnf("channel %s existed", id)
				_ = conn.WriteFrame(EIM.OpClose, []byte("channelId is repeated"))
				conn.Close()
				return
			}
			// 4. 创建一个channel对象, 并添加到连接管理中
			channel := EIM.NewChannel(id, conn)
			channel.SetWriteWait(s.options.writewait)
			channel.SetReadWait(s.options.readwait)
			s.Add(channel)
			log.Info("accept ", channel)
			// 5. 循环读取消息
			err = channel.Readloop(s.MessageListener)
			if err != nil {
				log.Info(err)
			}
			// 6. 如果Readloop方法返回了一个error, 说明连接已经断开, Server就需要把它从channelMap中删除
			s.Remove(channel.ID())
			// 7. 调用 (kim.StateListener).Disconnect(string) 把断开事件回调给业务层
			_ = s.Disconnect(channel.ID())
			channel.Close()
		}(rawconn)
	}
}

// Push 参数含义: id: channelId, data: 数据
func (s *Server) Push(id string, data []byte) error {
	ch, ok := s.ChannelMap.Get(id)
	if !ok {
		return errors.New("channel not found")
	}
	return ch.Push(data)
}

// Shutdown 关闭服务器
func (s *Server) Shutdown(ctx context.Context) error {
	log := logger.WithFields(logger.Fields{
		"module": "tcp.server",
		"id":     s.ServiceID(),
	})
	s.once.Do(func() {
		defer func() {
			log.Infoln("shutdown")
		}()
		// 关闭所有channel
		channels := s.ChannelMap.All()
		for _, ch := range channels {
			ch.Close()

			select {
			case <-ctx.Done():
				return
			default:
				continue
			}
		}
	})
	return nil
}

// defaultAcceptor 默认接收器
type defaultAcceptor struct{}

// Accept 回调, 交给上层处理认证等逻辑
func (a *defaultAcceptor) Accept(conn EIM.Conn, loginwait time.Duration) (string, error) {
	return ksuid.New().String(), nil
}
