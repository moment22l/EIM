package container

import (
	"EIM"
	"EIM/wire/pkt"
)

// Selector 用于选择一个服务
type Selector interface {
	Lookup(*pkt.Header, []EIM.Service) string
}
