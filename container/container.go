package container

import (
	"EIM"
	"EIM/logger"
	"EIM/naming"
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

//
const (
	stateUninitialized = iota
	stateInitialized
	stateStarted
	stateClosed
)

// Container 容器
type Container struct {
	sync.RWMutex
	Naming     naming.Naming
	Srv        EIM.Server
	state      uint32
	srvClients map[string]ClientMap
	selector   Selector
	dialer     EIM.Dialer
	deps       map[string]struct{}
}

var log = logger.WithField("module", "container")

// 默认单例容器
var c = &Container{
	state:    0,
	selector: &HashSelector{},
	deps:     make(map[string]struct{}),
}

// Default 返回默认容器
func Default() *Container {
	return c
}

// Init 初始化container
func Init(srv EIM.Server, deps ...string) error {
	// 检查是否已经初始化
	if !atomic.CompareAndSwapUint32(&c.state, stateUninitialized, stateInitialized) {
		return errors.New("has Initialized")
	}
	c.Srv = srv
	for _, dep := range deps {
		if _, ok := c.deps[dep]; ok {
			continue
		}
		c.deps[dep] = struct{}{}
	}
	log.WithField("func", "Init").Infof("srv %s:%s - deps %v", srv.ServiceID(), srv.ServiceName(), deps)
	c.srvClients = make(map[string]ClientMap, len(deps))
	return nil
}

func SetDialer(dialer EIM.Dialer) {
	c.dialer = dialer
}

func SetSelector(selector Selector) {
	c.selector = selector
}

func SetServiceNaming(nm naming.Naming) {
	c.Naming = nm
}

// Start 启动容器
func Start() error {
	if c.Naming == nil {
		return fmt.Errorf("naming is nil")
	}
	// 检查容器是否已启动
	if !atomic.CompareAndSwapUint32(&c.state, stateInitialized, stateStarted) {
		return errors.New("has started")
	}
	// 1. 启动server
	go func(server EIM.Server) {
		err := server.Start()
		if err != nil {
			log.Errorln(err)
		}
	}(c.Srv)
	// 2. 与依赖的服务进行连接
	for service := range c.deps {
		go func(service string) {
			err := ConnectToService(service)
			if err != nil {
				log.Errorln(err)
			}
		}(service)
	}
	// 3. 服务注册
	if c.Srv.PublicAddress() != "" && c.Srv.PublicPort() != 0 {
		err := c.Naming.Register(c.Srv)
		if err != nil {
			log.Errorln(err)
		}
	}
	// 4. 等待系统退出
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	log.Infoln("shutdown", <-c)
	// 5. 退出
	return shutdown()
}

// shutdown 退出容器
func shutdown() error {
	// 检查是否已退出
	if !atomic.CompareAndSwapUint32(&c.state, stateStarted, stateClosed) {
		return errors.New("has closed")
	}

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*10)
	defer cancel()
	// 1. 优雅退出服务器
	err := c.Srv.Shutdown(ctx)
	if err != nil {
		log.Error(err)
	}
	// 2. 注销服务
	err = c.Naming.Deregister(c.Srv.ServiceID())
	if err != nil {
		log.Warn(err)
	}
	// 3. 退订服务变更
	for dep := range c.deps {
		_ = c.Naming.Unsubscribe(dep)
	}
	log.Infoln("shutdown")
	return nil
}

// ConnectToService 连接服务
func ConnectToService(serviceName string) error {
	// TODO:implement me
	return nil
}
