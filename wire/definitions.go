package wire

type Magic [4]byte

var (
	MagicLoginPkt = Magic{0xc3, 0x11, 0xa3, 0x65}
	MagicBasicPkt = Magic{0xc3, 0x15, 0xa7, 0x65}
)

type protocol string

const (
	ProtocolTCP       protocol = "tcp"
	ProtocolWebsocket protocol = "websocket"
)

// Meta key of a packet
const (
	// MetaDestServer 表示Meta中的value为消息将要送达的网关的ServiceName(消息抵达的服务)
	MetaDestServer = "dest.server"
	// MetaDestChannels 表示Meta中的value为消息将要送达的channels(消息接收方)
	MetaDestChannels = "dest.channels"
)

// Service Name 统一的服务名称
const (
	SNWGateway = "wgateway"
	SNTGateway = "tgateway"
	SNLogin    = "login"
	SNChat     = "chat"
)
