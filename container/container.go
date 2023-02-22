package container

import (
	"EIM"
	"EIM/logger"
	"EIM/naming"
	"EIM/tcp"
	"EIM/wire"
	"EIM/wire/pkt"
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

const (
	stateUninitialized = iota
	stateInitialized
	stateStarted
	stateClosed
)

const (
	StateYoung = "young"
	StateAdult = "adult"
)

const (
	KeyServiceState = "service_state"
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
	clients := NewClients(10)
	c.srvClients[serviceName] = clients
	// watch服务的新增
	delay := time.Second * 10
	err := c.Naming.Subscribe(serviceName, func(services []EIM.ServiceRegistration) {
		for _, service := range services {
			if _, ok := clients.Get(service.ServiceID()); ok {
				continue
			}
			log.WithField("func", "connectToService").Infof("Watch a new service: %v", service)
			service.GetMeta()[KeyServiceState] = StateYoung
			go func(service EIM.ServiceRegistration) {
				time.Sleep(delay)
				service.GetMeta()[KeyServiceState] = StateAdult
			}(service)

			_, err := buildClient(clients, service)
			if err != nil {
				logger.Warn(err)
			}
		}
	})
	if err != nil {
		return err
	}
	// 查询已经存在的服务
	services, err := c.Naming.Find(serviceName)
	if err != nil {
		return err
	}
	log.Info("find service", services)
	for _, service := range services {
		service.GetMeta()[KeyServiceState] = StateAdult
		_, err := buildClient(clients, service)
		if err != nil {
			logger.Warn(err)
		}
	}
	return nil
}

// buildClient 使clients与service进行连接
func buildClient(clients ClientMap, service EIM.ServiceRegistration) (EIM.Client, error) {
	c.Lock()
	defer c.Unlock()
	var (
		id   = service.ServiceID()
		name = service.ServiceName()
		meta = service.GetMeta()
	)
	// 1. 检查连接是否已经存在
	if _, ok := clients.Get(id); ok {
		return nil, nil
	}
	// 2. 检查协议是否为TCP, 服务之间只允许使用TCP协议
	if service.GetProtocol() != string(wire.ProtocolTCP) {
		return nil, fmt.Errorf("unexpected service Protocol: %s", service.GetProtocol())
	}
	// 3. 构建客户端并连接
	cli := tcp.NewClientWithProps(id, name, meta, tcp.ClientOptions{
		Heartbeat: EIM.DefaultHeartbeat,
		ReadWait:  EIM.DefaultReadWait,
		WriteWait: EIM.DefaultWriteWait,
	})
	if c.dialer == nil {
		return nil, fmt.Errorf("dialer is nil")
	}
	cli.SetDialer(c.dialer)
	err := cli.Connect(service.DialURL())
	if err != nil {
		return nil, err
	}
	// 4. 读取消息
	go func(cli EIM.Client) {
		err := readLoop(cli)
		if err != nil {
			log.Debug(err)
		}
		clients.Remove(id)
		cli.Close()
	}(cli)
	// 5. 添加到客户端中
	clients.Add(cli)
	return cli, nil
}

// readLoop 循环接收信息
func readLoop(cli EIM.Client) error {
	log := logger.WithFields(logger.Fields{
		"module": "container",
		"func":   "readLoop",
	})
	log.Infof("readLoop started of %s %s", cli.ID(), cli.Name())
	for {
		frame, err := cli.Read()
		if err != nil {
			return err
		}
		if frame.GetOpCode() != EIM.OpBinary {
			continue
		}
		buf := bytes.NewBuffer(frame.GetPayload())

		p, err := pkt.MustReadLoginPkt(buf)
		if err != nil {
			log.Info(err)
			continue
		}
		err = PushMessage(p)
		if err != nil {
			log.Info(err)
		}
	}
}

// Push 供上层业务调用, 用于将消息发送给网关
func Push(server string, p *pkt.LoginPkt) error {
	p.AddStringMeta(wire.MetaDestServer, server)
	return c.Srv.Push(server, pkt.Marshal(p))
}

// PushMessage 将从Client中收到的消息通过Server找到对应channel并发送到客户端
func PushMessage(p *pkt.LoginPkt) error {
	server, _ := p.GetMeta(wire.MetaDestServer)
	if server != c.Srv.ServiceID() {
		return fmt.Errorf("dest_server is incorrect, %s != %s", server, c.Srv.ServiceID())
	}
	channels, ok := p.GetMeta(wire.MetaDestChannels)
	if !ok {
		return fmt.Errorf("dest_channels is nil")
	}

	channelsIds := strings.Split(channels.(string), ",")
	p.DelMeta(wire.MetaDestServer)
	p.DelMeta(wire.MetaDestChannels)
	payload := pkt.Marshal(p)
	log.Debugf("Push to %v %v", channelsIds, p)

	for _, channel := range channelsIds {
		err := c.Srv.Push(channel, payload)
		if err != nil {
			log.Debug(err)
		}
	}
	return nil
}

// Forward 消息上行, 下游服务发送消息到上游服务
func Forward(serviceName string, p *pkt.LoginPkt) error {
	if p == nil {
		return errors.New("packet is nil")
	}
	if p.Command == "" {
		return errors.New("command is empty in packet")
	}
	if p.ChannelId == "" {
		return errors.New("channelId is empty in packet")
	}
	return ForwardWithSelector(serviceName, p, c.selector)
}

// ForwardWithSelector 可以指定一个Selector来推送消息到服务的指定节点
func ForwardWithSelector(serviceName string, p *pkt.LoginPkt, selector Selector) error {
	cli, err := lookup(serviceName, &p.Header, selector)
	if err != nil {
		return err
	}
	// 加一个tag到packet中
	p.AddStringMeta(wire.MetaDestServer, c.Srv.ServiceID())
	log.Debugf("forward message to %v with %s", cli.ID(), &p.Header)

	return cli.Send(pkt.Marshal(p))
}

// lookup 根据服务名查找一个可靠服务
func lookup(serviceName string, header *pkt.Header, selector Selector) (EIM.Client, error) {
	clients, ok := c.srvClients[serviceName]
	if !ok {
		return nil, fmt.Errorf("service %s not found", serviceName)
	}
	// 只获取状态为StateAdult的服务
	srvs := clients.Services(KeyServiceState, StateAdult)
	if len(srvs) == 0 {
		return nil, fmt.Errorf("no services found for %s", serviceName)
	}
	id := selector.Lookup(header, srvs)
	if cli, ok := clients.Get(id); ok {
		return cli, nil
	}
	return nil, fmt.Errorf("no client found")
}
