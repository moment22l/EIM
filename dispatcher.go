package EIM

import "EIM/wire/pkt"

// Dispatcher 消息分发器, 向网关gateway中的channels两个连接推送一条消息LogicPkt消息
type Dispatcher interface {
	Push(gateway string, channels []string, pkt *pkt.LoginPkt) error
}
