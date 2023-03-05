package server

import (
	"EIM"
	"EIM/container"
	"EIM/logger"
	"EIM/naming"
	"EIM/naming/consul"
	"EIM/services/server/conf"
	"EIM/services/server/handler"
	"EIM/services/server/serv"
	"EIM/storage"
	"EIM/tcp"
	"EIM/wire"
	"context"
	"time"
)

type ServerStartOptions struct {
	config      string
	serviceName string
}

// RunServerStart 启动服务
func RunServerStart(ctx context.Context, opts *ServerStartOptions, version string) error {
	// 获取配置
	config, err := conf.Init(opts.config)
	if err != nil {
		return err
	}
	// 初始化logger
	_ = logger.Init(logger.Settings{
		Level: "trace",
	})
	// 初始化Router
	r := EIM.NewRouter()
	// login
	loginHandler := handler.NewLoginHandler()
	r.Handle(wire.CommandLoginSignIn, loginHandler.DoSysLogin)
	r.Handle(wire.CommandLoginSignOut, loginHandler.DoSysLogout)
	// 初始化redis
	redis, err := conf.InitRedis(config.RedisAddr, "")
	if err != nil {
		return err
	}
	// 初始化会话管理
	cache := storage.NewRedisStorage(redis)
	servHandler := serv.NewServHandler(r, cache)
	service := &naming.DefaultService{
		Id:        config.ServiceID,
		Name:      opts.serviceName,
		Address:   config.PublicAddress,
		Port:      config.PublicPort,
		Protocol:  string(wire.ProtocolTCP),
		Namespace: config.Namespace,
		Tags:      config.Tags,
	}
	// 初始化server及注册监听器
	srv := tcp.NewServer(config.Listen, service)
	srv.SetReadWait(time.Minute * 2)
	srv.SetAcceptor(servHandler)
	srv.SetMessageListener(servHandler)
	srv.SetStateListener(servHandler)
	// 初始化container
	err = container.Init(srv)
	if err != nil {
		return err
	}
	// 初始化naming
	ns, err := consul.NewNaming(config.ConsulURL)
	if err != nil {
		return err
	}
	container.SetServiceNaming(ns)
	// 启动container
	return container.Start()
}
