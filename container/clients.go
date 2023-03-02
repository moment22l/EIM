package container

import (
	"EIM"
	"EIM/logger"
	"sync"
)

// ClientMap 接口
type ClientMap interface {
	Add(client EIM.Client)
	Remove(id string)
	Get(id string) (client EIM.Client, ok bool)
	Services(kvs ...string) []EIM.Service
}

// ClientsImpl 包含存储着clients的Map
type ClientsImpl struct {
	clients *sync.Map
}

// Add 向c.clients中添加client
func (c *ClientsImpl) Add(client EIM.Client) {
	if client.ServiceID() == "" {
		logger.WithFields(logger.Fields{
			"module": "ClientsImpl",
		}).Error("client id is required")
	}
	c.clients.Store(client.ServiceID(), client)
}

// Remove 删除c.clients中对应id的client
func (c *ClientsImpl) Remove(id string) {
	c.clients.Delete(id)
}

// Get 获取对应id的client
func (c *ClientsImpl) Get(id string) (client EIM.Client, ok bool) {
	if id == "" {
		logger.WithFields(logger.Fields{
			"module": "ClientsImpl",
		}).Error("client id is required")
	}

	if value, ok := c.clients.Load(id); ok {
		client = value.(EIM.Client)
		return client, true
	}
	return nil, false
}

// Services 返回服务列表，可以传入键值对
func (c *ClientsImpl) Services(kvs ...string) []EIM.Service {
	kvLen := len(kvs)
	if kvLen != 0 && kvLen != 2 {
		return nil
	}
	arr := make([]EIM.Service, 0)
	c.clients.Range(func(key, value any) bool {
		service := value.(EIM.Service)
		if kvLen > 0 && service.GetMeta()[kvs[0]] != kvs[1] {
			return true
		}
		arr = append(arr, service)
		return true
	})
	return arr
}

// NewClients 创建新的ClientsImp并返回(返回类型为接口ClientMap)
func NewClients() ClientMap {
	return &ClientsImpl{
		clients: new(sync.Map),
	}
}
