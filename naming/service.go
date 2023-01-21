package naming

import "fmt"

// ServiceRegistration service接口定义
type ServiceRegistration interface {
	ServiceID() string
	ServiceName() string
	// ip or domain
	PublicAddress() string
	PublicPort() int
	DialURL() string
	GetProtocol() string
	GetNamespace() string
	GetTags() []string
	GetMeta() map[string]string
	// SetTags(tags []string)
	// SetMeta(map[string]string)
	String() string
}

// DefaultService 实现Service接口
type DefaultService struct {
	Id        string
	Name      string
	Address   string
	Port      int
	Protocol  string
	Namespace string
	Tags      []string
	Meta      map[string]string
}

// NewEntry 新建条目
func NewEntry(id, name, protocol string, address string, port int) ServiceRegistration {
	return &DefaultService{
		Id:       id,
		Name:     name,
		Address:  address,
		Port:     port,
		Protocol: protocol,
	}
}

// ServiceID 返回e的Id
func (e *DefaultService) ServiceID() string {
	return e.Id
}

// ServiceName 返回e的Name
func (e *DefaultService) ServiceName() string {
	return e.Name
}

// PublicAddress 返回e的address
func (e *DefaultService) PublicAddress() string {
	return e.Address
}

// PublicPort 返回e的port
func (e *DefaultService) PublicPort() int {
	return e.Port
}

// DialURL 返回e的拨号URL
func (e *DefaultService) DialURL() string {
	if e.Protocol == "tcp" {
		return fmt.Sprintf("%s:%d", e.Address, e.Port)
	}
	return fmt.Sprintf("%s://%s:%d", e.Protocol, e.Address, e.Port)
}

// GetProtocol 返回e的Protocol
func (e *DefaultService) GetProtocol() string {
	return e.Protocol
}

// GetNamespace 返回e的Namespace
func (e *DefaultService) GetNamespace() string {
	return e.Namespace
}

// GetTags 返回e的Tags
func (e *DefaultService) GetTags() []string {
	return e.Tags
}

// GetMeta 返回e的Meta
func (e *DefaultService) GetMeta() map[string]string {
	return e.Meta
}

// String 返回e的所有相关信息
func (e *DefaultService) String() string {
	return fmt.Sprintf("Id:%s,Name:%s,Address:%s,Port:%d,Namespace:%s,Tags:%v,Meta:%v",
		e.Id, e.Name, e.Address, e.Port, e.Namespace, e.Tags, e.Meta)
}

// // SetTags 设置e的Tags
// func (e *DefaultService) SetTags(tags []string) {
// 	e.Tags = tags
// }
//
// // SetMeta 设置e的Meta
// func (e *DefaultService) SetMeta(meta map[string]string) {
// 	e.Meta = meta
// }
