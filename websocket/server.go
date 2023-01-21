package websocket

import (
	"EIM"
	"EIM/logger"
	"EIM/naming"
	"context"
	"errors"
	"fmt"
	"github.com/gobwas/ws"
	"github.com/segmentio/ksuid"
	"net/http"
	"sync"
	"time"
)

// ServerOptions Server的超时参数
type ServerOptions struct {
	loginwait time.Duration // 登录超时
	readwait  time.Duration // 读超时
	writewait time.Duration // 写超时
}

// Server websocket的Server实现
type Server struct {
	listen string
	naming.ServiceRegistration
	EIM.MessageListener
	EIM.StateListener
	EIM.Acceptor
	EIM.ChannelMap
	once    sync.Once
	options ServerOptions
}

// NewServer 创建一个新Server
func NewServer(listen string, service naming.ServiceRegistration) EIM.Server {
	return &Server{
		listen:              listen,
		ServiceRegistration: service,
		options: ServerOptions{
			loginwait: EIM.DefaultLoginWait,
			readwait:  EIM.DefaultReadWait,
			writewait: EIM.DefaultWriteWait,
		},
	}
}

// defaultAcceptor 实现了Acceptor接口
type defaultAcceptor struct{}

// Accept 回调, 交给上层处理认证等逻辑
func (d *defaultAcceptor) Accept(conn EIM.Conn, loginwait time.Duration) (string, error) {
	return ksuid.New().String(), nil
}

// Start websocket的Server的主要逻辑部分, 开启服务器
func (s *Server) Start() error {
	mux := http.NewServeMux()
	log := logger.WithFields(logger.Fields{
		"module": "ws.server",
		"listen": s.listen,
		"id":     s.ServiceID(),
	})

	if s.Acceptor == nil {
		s.Acceptor = new(defaultAcceptor)
	}
	if s.StateListener == nil {
		return fmt.Errorf("StateListener is nil")
	}
	if s.ChannelMap == nil {
		s.ChannelMap = EIM.NewChannels(100)
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// 1. 升级连接
		rawconn, _, _, err := ws.UpgradeHTTP(r, w)
		if err != nil {
			resp(w, http.StatusBadRequest, err.Error())
			return
		}
		// 2. 包装conn
		conn := NewConn(rawconn)
		// 3. 回调到上层业务完成权限认证之类的逻辑处理
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
		// 4. 添加channel到kim.ChannelMap连接管理器中
		channel := EIM.NewChannel(id, conn)
		channel.SetWriteWait(s.options.writewait)
		channel.SetReadWait(s.options.readwait)
		s.Add(channel)

		// 5. 开启一个goroutine中循环读取消息
		go func(ch EIM.Channel) {
			err = ch.Readloop(s.MessageListener)
			if err != nil {
				log.Info(err)
			}
			// 6. 退出
			s.Remove(ch.ID())
			err = s.Disconnect(ch.ID())
			if err != nil {
				log.Warn(err)
			}
			err = ch.Close()
			if err != nil {
				return
			}
		}(channel)
	})
	log.Infoln("started")
	return http.ListenAndServe(s.listen, mux)
}

// Push 推送一个消息到channel
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
		"module": "ws.server",
		"id":     s.ServiceID(),
	})
	s.once.Do(func() {
		defer func() {
			log.Infoln("shutdown")
		}()
		// 关闭所有channels
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

// resp 将body写入w
func resp(w http.ResponseWriter, code int, body string) {
	w.WriteHeader(code)
	if body != "" {
		_, _ = w.Write([]byte(body))
	}
	logger.Warnf("response with code:%d %s", code, body)
}
