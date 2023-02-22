package container

import (
	"EIM"
	"sync"
)

type ClientMap interface {
	Add(client EIM.Client)
	Remove(id string)
	Get(id string) (client EIM.Client, ok bool)
	Services(kvs ...string) []EIM.Service
}

type ClientsImp struct {
	clients *sync.Map
}

func (c *ClientsImp) Add(client EIM.Client) {
	//TODO implement me
	panic("implement me")
}

func (c *ClientsImp) Remove(id string) {
	//TODO implement me
	panic("implement me")
}

func (c *ClientsImp) Get(id string) (client EIM.Client, ok bool) {
	//TODO implement me
	panic("implement me")
}

func (c *ClientsImp) Services(kvs ...string) []EIM.Service {
	//TODO implement me
	panic("implement me")
}

func NewClients(num int) ClientMap {
	return &ClientsImp{
		clients: new(sync.Map),
	}
}
