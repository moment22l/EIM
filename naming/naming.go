package naming

import "errors"

// errors
var (
	ErrNotFound = errors.New("service not found")
)

// Naming 接口 定义关于naming服务的方法
type Naming interface {
	Find(serviceName string) ([]ServiceRegistration, error)
	Remove(serviceName, serviceID string) error
	Register(ServiceRegistration) error
	Deregister(serviceID string) error
	// Get(namespace string, id string) (ServiceRegistration, error)
}
