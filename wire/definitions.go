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
