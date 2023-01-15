package websocket

import (
	"EIM"
	"context"
	"fmt"
	"github.com/gobwas/ws"
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

func (d *defaultAcceptor) Accept(conn EIM.Conn, loginwait time.Duration) (string, error) {
	//TODO implement me
	panic("implement me")
}

// Start websocket的Server的主要逻辑部分
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

func (s *Server) Push(s2 string, bytes []byte) error {
	//TODO implement me
	panic("implement me")
}

func (s *Server) Shutdown(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}

func (s *Server) SetAcceptor(acceptor EIM.Acceptor) {
	s.Acceptor = acceptor
}

func (s *Server) SetMessageListener(listener EIM.MessageListener) {
	s.MessageListener = listener
}

func (s *Server) SetStateListener(listener EIM.StateListener) {
	s.StateListener = listener
}

func (s *Server) SetChannelMap(channelMap EIM.ChannelMap) {
	s.ChannelMap = channelMap
}

func (s *Server) SetReadWait(readwait time.Duration) {
	s.options.readwait = readwait
}

func resp(w http.ResponseWriter, code int, body string) {
	w.WriteHeader(code)
	if body != "" {
		_, _ = w.Write([]byte(body))
	}
	logger.Warnf("response with code:%d %s", code, body)
}
