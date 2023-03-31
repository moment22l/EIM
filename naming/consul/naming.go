package consul

import (
	"EIM"
	"EIM/logger"
	"EIM/naming"
	"errors"
	"fmt"
	"sync"

	"github.com/hashicorp/consul/api"
)

const (
	KeyProtocol  = "protocol"
	KeyHealthURL = "health_url"
)

type Watch struct {
	Service   string
	Callback  func(services []EIM.ServiceRegistration)
	WaitIndex uint64
	Quit      chan struct{}
}

// Naming 实现了naming.Naming接口
type Naming struct {
	sync.RWMutex
	cli    *api.Client
	watchs map[string]*Watch
}

// NewNaming 创建一个新Naming并返回
func NewNaming(consulUrl string) (naming.Naming, error) {
	conf := api.DefaultConfig()
	conf.Address = consulUrl
	cli, err := api.NewClient(conf)
	if err != nil {
		return nil, err
	}
	n := &Naming{
		cli:    cli,
		watchs: make(map[string]*Watch, 1),
	}
	return n, nil
}

// Register 服务注册
func (n *Naming) Register(service EIM.ServiceRegistration) error {
	reg := &api.AgentServiceRegistration{
		ID:      service.ServiceID(),
		Name:    service.ServiceName(),
		Address: service.PublicAddress(),
		Port:    service.PublicPort(),
		Tags:    service.GetTags(),
		Meta:    service.GetMeta(),
	}
	if reg.Meta == nil {
		reg.Meta = make(map[string]string)
	}
	reg.Meta[KeyProtocol] = service.GetProtocol()

	// 健康检查
	healthURL := service.GetMeta()[KeyHealthURL]
	if healthURL != "" {
		check := new(api.AgentServiceCheck)
		check.CheckID = fmt.Sprintf("%s_normal", service.ServiceID())
		check.HTTP = healthURL
		check.Timeout = "1s"
		check.Interval = "10s"
		check.DeregisterCriticalServiceAfter = "20s"
		reg.Check = check
	}
	return n.cli.Agent().ServiceRegister(reg)
}

// Deregister 服务取消注册
func (n *Naming) Deregister(serviceID string) error {
	return n.cli.Agent().ServiceDeregister(serviceID)
}

// Find 服务发现, 可添加tags
func (n *Naming) Find(serviceName string, tags ...string) ([]EIM.ServiceRegistration, error) {
	services, _, err := n.load(serviceName, 0, tags...)
	if err != nil {
		return nil, err
	}
	return services, nil
}

// load 刷新服务注册
func (n *Naming) load(serviceName string, waitIndex uint64,
	tags ...string) ([]EIM.ServiceRegistration, *api.QueryMeta, error) {
	opts := &api.QueryOptions{
		UseCache:  true,
		MaxAge:    0,
		WaitIndex: waitIndex,
	}
	catalogServices, meta, err := n.cli.Catalog().ServiceMultipleTags(serviceName, tags, opts)
	if err != nil {
		return nil, meta, err
	}
	services := make([]EIM.ServiceRegistration, 0, len(catalogServices))
	for _, s := range catalogServices {
		if s.Checks.AggregatedStatus() != api.HealthPassing {
			logger.Debugf("load service: id:%s name:%s %s:%d Status:%s", s.ServiceID, s.ServiceName,
				s.ServiceAddress, s.ServicePort, s.Checks.AggregatedStatus())
			continue
		}
		services = append(services, &naming.DefaultService{
			Id:       s.ServiceID,
			Name:     s.ServiceName,
			Address:  s.ServiceAddress,
			Port:     s.ServicePort,
			Protocol: s.ServiceMeta[KeyProtocol],
			Tags:     s.ServiceTags,
			Meta:     s.ServiceMeta,
		})
	}
	return services, meta, nil
}

// Subscribe 订阅服务
func (n *Naming) Subscribe(serviceName string, callback func(services []EIM.ServiceRegistration)) error {
	n.Lock()
	defer n.Unlock()
	if _, ok := n.watchs[serviceName]; ok {
		return errors.New("service has already been registered")
	}
	w := &Watch{
		Service:  serviceName,
		Callback: callback,
		Quit:     make(chan struct{}),
	}
	n.watchs[serviceName] = w

	go n.watch(w)
	return nil
}

// watch 用于订阅服务
func (n *Naming) watch(w *Watch) {
	stopped := false
	var doWatch = func(service string, callback func([]EIM.ServiceRegistration)) {
		services, meta, err := n.load(service, w.WaitIndex)
		if err != nil {
			logger.Warn(err)
			return
		}
		select {
		case <-w.Quit:
			stopped = true
			logger.Info("watch %s stopped", w.Service)
			return
		default:

		}
		w.WaitIndex = meta.LastIndex
		if callback != nil {
			callback(services)
		}
	}

	// 初始化w.WaitIndex
	doWatch(w.Service, nil)
	for !stopped {
		doWatch(w.Service, w.Callback)
	}
}

// Unsubscribe 取消订阅服务
func (n *Naming) Unsubscribe(serviceName string) error {
	n.Lock()
	defer n.Unlock()
	watch, ok := n.watchs[serviceName]
	delete(n.watchs, serviceName)
	if ok {
		close(watch.Quit)
	}
	return nil
}
