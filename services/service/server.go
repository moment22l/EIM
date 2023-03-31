package service

import (
	"EIM/logger"
	"EIM/naming"
	"EIM/naming/consul"
	"EIM/services/service/conf"
	"EIM/services/service/database"
	"EIM/services/service/handler"
	"EIM/wire"
	"context"
	"fmt"
	"gorm.io/gorm"
	"hash/crc32"

	"github.com/kataras/iris/v12"
	"github.com/spf13/cobra"
)

type ServerStartOptions struct {
	config string
}

func NewServerStartCmd(ctx context.Context, version string) *cobra.Command {
	opts := &ServerStartOptions{}

	cmd := &cobra.Command{
		Use:   "royal",
		Short: "start a rpc service",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunServerStart(ctx, opts, version)
		},
	}
	cmd.PersistentFlags().StringVarP(&opts.config, "config", "c", "./service/conf.yaml", "Config file")
	return cmd
}

func RunServerStart(ctx context.Context, opts *ServerStartOptions, version string) error {
	config, err := conf.Init(opts.config)
	if err != nil {
		return err
	}
	_ = logger.Init(logger.Settings{
		Level:    config.LogLevel,
		Filename: "./data/royal.log",
	})

	// 初始化DB
	var (
		baseDB    *gorm.DB
		messageDB *gorm.DB
	)
	baseDB, err = database.InitDB(config.Driver, config.BaseDB)
	if err != nil {
		return err
	}
	messageDB, err = database.InitDB(config.Driver, config.MessageDB)
	if err != nil {
		return err
	}
	// 迁移对应模型
	_ = baseDB.AutoMigrate(&database.Group{}, &database.GroupMember{})
	_ = messageDB.AutoMigrate(&database.MessageIndex{}, &database.MessageContent{})

	// 处理NodeID为0的情况
	if config.NodeID == 0 {
		config.NodeID = int64(HashCode(config.ServiceID))
	}
	idGen, err := database.NewIDGenerator(config.NodeID)
	if err != nil {
		return err
	}
	// 初始化redis
	redis, err := conf.InitRedis(config.RedisAddrs, "")
	if err != nil {
		return err
	}
	// 初始化注册中心consul
	ns, err := consul.NewNaming(config.ConsulURL)
	if err != nil {
		return err
	}
	err = ns.Register(&naming.DefaultService{
		Id:       config.ServiceID,
		Name:     wire.SNService,
		Address:  config.PublicAddress,
		Port:     config.PublicPort,
		Protocol: "http",
		Tags:     config.Tags,
		Meta: map[string]string{
			consul.KeyHealthURL: fmt.Sprintf("http://%s:%d/health", config.PublicAddress, config.PublicPort),
		},
	})
	if err != nil {
		return err
	}
	defer func() {
		_ = ns.Deregister(config.ServiceID)
	}()
	serviceHandler := handler.ServiceHandler{
		BaseDB:    baseDB,
		MessageDB: messageDB,
		Cache:     redis,
		IDGen:     idGen,
	}

	ac := conf.MakeAccessLog()
	defer ac.Close()

	app := newApp(&serviceHandler)
	app.UseRouter(ac.Handler)
	app.UseRouter(setAllowedResponse)

	// Start server
	return app.Listen(config.Listen, iris.WithOptimizations)
}

func newApp(handler *handler.ServiceHandler) *iris.Application {
	app := iris.Default()

	app.Get("/health", func(ctx iris.Context) {
		_, _ = ctx.WriteString("ok")
	})
	messageAPI := app.Party("/api/:app/message")
	{
		messageAPI.Post("/user", handler.InsertUserMessage)
		messageAPI.Post("/group", handler.InsertGroupMessage)
		messageAPI.Post("/ack", handler.MessageAck)
	}

	groupAPI := app.Party("/api/:app/group")
	{
		groupAPI.Get("/:id", handler.GroupGet)
		groupAPI.Post("", handler.GroupCreate)
		groupAPI.Post("/member", handler.GroupJoin)
		groupAPI.Delete("/member", handler.GroupQuit)
		groupAPI.Get("/members/:id", handler.GroupMembers)
	}

	offlineAPI := app.Party("/api/:app/offline")
	{
		offlineAPI.Use(iris.Compression)
		offlineAPI.Post("/index", handler.GetOfflineMessageIndex)
		offlineAPI.Post("/content", handler.GetOfflineMessageContent)
	}
	return app
}

func setAllowedResponse(ctx iris.Context) {
	// Indicate that the Server can send JSON and Protobuf for this request.
	ctx.Negotiation().JSON().Protobuf()

	// If client is missing an "Accept: " header then default it to JSON.
	ctx.Negotiation().Accept.JSON()

	ctx.Next()
}

func HashCode(key string) uint32 {
	hash32 := crc32.NewIEEE()
	_, err := hash32.Write([]byte(key))
	if err != nil {
		return 0
	}
	return hash32.Sum32() % 1000
}
