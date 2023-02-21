package container

import (
	"EIM"
	"EIM/wire/pkt"
	"hash/crc32"
)

// HashSelector 实现了Selector接口, 容器的一个默认selector
type HashSelector struct {
}

// Lookup 找到对应服务并返回其ID
func (s *HashSelector) Lookup(header *pkt.Header, services []EIM.Service) string {
	ll := len(services)
	code := HashCode(header.ChannelId)
	return services[code%ll].ServiceID()
}

// HashCode 利用crc32算法得到一个数字
func HashCode(key string) int {
	hash32 := crc32.NewIEEE()
	hash32.Write([]byte(key))
	return int(hash32.Sum32())
}
