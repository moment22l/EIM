package container

import "EIM"

type ClientMap interface {
	Add(client EIM.Client)
	Remove(id string)
	Get(id string) (client EIM.Client, ok bool)
	Services(kvs ...string) []EIM.Service
}
