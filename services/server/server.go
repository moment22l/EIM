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
	"EIM/services/server/service"
	"EIM/storage"
	"EIM/tcp"
	"EIM/wire"
	"context"
	"fmt"
	"github.com/go-resty/resty/v2"
	"strings"
	"time"

	"github.com/spf13/cobra"
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
		Level:    config.LogLevel,
		Filename: "./data/server.log",
	})
	var groupService service.Group
	var messageService service.Message
	if strings.TrimSpace(config.RoyalURL) != "" {
		groupService = service.NewGroupService(config.RoyalURL)
		messageService = service.NewMessageService(config.RoyalURL)
	} else {
		srv := &resty.SRVRecord{
			Service: "consul",
			Domain:  wire.SNService,
		}
		groupService = service.NewGroupServiceWithSRV("http", srv)
		messageService = service.NewMessageServiceWithSRV("http", srv)
	}
	// 初始化Router
	r := EIM.NewRouter()
	// login
	loginHandler := handler.NewLoginHandler()
	r.Handle(wire.CommandLoginSignIn, loginHandler.DoSysLogin)
	r.Handle(wire.CommandLoginSignOut, loginHandler.DoSysLogout)
	// talk
	chatHandler := handler.NewChatHandler(messageService, groupService)
	r.Handle(wire.CommandChatUserTalk, chatHandler.DoUserTalk)
	r.Handle(wire.CommandChatGroupTalk, chatHandler.DoGroupTalk)
	// group
	groupHandler := handler.NewGroupHandler(groupService)
	r.Handle(wire.CommandGroupCreate, groupHandler.DoCreate)
	r.Handle(wire.CommandGroupJoin, groupHandler.DoJoin)
	r.Handle(wire.CommandGroupQuit, groupHandler.DoQuit)
	r.Handle(wire.CommandGroupDetail, groupHandler.DoDetail)
	// offline
	offlineHandler := handler.NewOfflineHandler(messageService)
	r.Handle(wire.CommandOfflineIndex, offlineHandler.DoSyncIndex)
	r.Handle(wire.CommandOfflineContent, offlineHandler.DoSyncContent)
	// 初始化redis
	redis, err := conf.InitRedis(config.RedisAddrs, "")
	if err != nil {
		return err
	}
	// 初始化会话管理
	cache := storage.NewRedisStorage(redis)
	servHandler := serv.NewServHandler(r, cache)
	meta := make(map[string]string)
	meta[consul.KeyHealthURL] = fmt.Sprintf("http://%s:%d/health", config.PublicAddress, config.MonitorPort)
	service := &naming.DefaultService{
		Id:       config.ServiceID,
		Name:     opts.serviceName,
		Address:  config.PublicAddress,
		Port:     config.PublicPort,
		Protocol: string(wire.ProtocolTCP),
		Tags:     config.Tags,
		Meta:     meta,
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

func NewServerStartCmd(ctx context.Context, version string) *cobra.Command {
	opts := &ServerStartOptions{}

	cmd := &cobra.Command{
		Use:   "server",
		Short: "start a server",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunServerStart(ctx, opts, version)
		},
	}
	cmd.PersistentFlags().StringVarP(&opts.config, "config", "c", "./server/conf.yaml", "Config file")
	cmd.PersistentFlags().StringVarP(&opts.serviceName, "serviceName", "s", "chat", "defined a service name(login or chat)")
	return cmd
}
