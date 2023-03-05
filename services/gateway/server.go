package gateway

import (
	"EIM"
	"EIM/container"
	"EIM/logger"
	"EIM/naming"
	"EIM/naming/consul"
	"EIM/services/gateway/conf"
	"EIM/services/gateway/serv"
	"EIM/websocket"
	"EIM/wire"
	"context"
	"time"
)

type ServerStartOptions struct {
	config   string
	protocol string
}

// RunServerStart 启动网关
func RunServerStart(ctx context.Context, opts *ServerStartOptions, version string) error {
	// 获取配置
	config, err := conf.Init(opts.config)
	if err != nil {
		return err
	}
	// 初始化日志
	_ = logger.Init(logger.Settings{
		Level: "trace",
	})
	// 初始化handler
	handler := &serv.Handler{
		ServiceID: config.ServiceID,
	}
	// 初始化server
	var srv EIM.Server
	service := &naming.DefaultService{
		Id:        config.ServiceID,
		Name:      config.ServiceName,
		Address:   config.PublicAddress,
		Port:      config.PublicPort,
		Protocol:  opts.protocol,
		Namespace: config.Namespace,
		Tags:      config.Tags,
	}
	if opts.protocol == "ws" {
		srv = websocket.NewServer(config.Listen, service)
	}
	// 注册监听器
	srv.SetReadWait(time.Minute * 2)
	srv.SetStateListener(handler)
	srv.SetMessageListener(handler)
	srv.SetAcceptor(handler)
	// 初始化container
	_ = container.Init(srv, wire.SNChat, wire.SNLogin)
	// 初始化naming
	ns, err := consul.NewNaming(config.ConsulURL)
	if err != nil {
		return err
	}
	container.SetServiceNaming(ns)
	container.SetDialer(serv.NewDialer(config.ServiceID))
	// 启动容器
	return container.Start()
}
