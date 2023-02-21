package naming

import (
	"EIM"
	"errors"
)

// errors
var (
	ErrNotFound = errors.New("service not found")
)

// Naming 接口 定义关于naming服务的方法
type Naming interface {
	Find(serviceName string) ([]EIM.ServiceRegistration, error)
	Subscribe(serviceName string, callback func(services []EIM.ServiceRegistration)) error
	Unsubscribe(serviceName string) error
	Register(EIM.ServiceRegistration) error
	Deregister(serviceID string) error
	// Get(namespace string, id string) (ServiceRegistration, error)
}
