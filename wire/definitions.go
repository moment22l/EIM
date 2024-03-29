package wire

import "time"

type Magic [4]byte

var (
	MagicLogicPkt = Magic{0xc3, 0x11, 0xa3, 0x65}
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
	SNService  = "service"
)

type ServiceID string

type SessionID string

// Command的类型
const (
	// login
	CommandLoginSignIn  = "login.signin"
	CommandLoginSignOut = "login.signout"

	// chat
	CommandChatUserTalk  = "chat.user.talk"
	CommandChatGroupTalk = "chat.group.talk"
	CommandChatTalkAck   = "chat.talk.ack"

	// 离线
	CommandOfflineIndex   = "chat.offline.index"
	CommandOfflineContent = "chat.offline.content"

	// 群管理
	CommandGroupCreate  = "chat.group.create"
	CommandGroupJoin    = "chat.group.join"
	CommandGroupQuit    = "chat.group.quit"
	CommandGroupMembers = "chat.group.members"
	CommandGroupDetail  = "chat.group.detail"
)

const (
	MessageTypeText  = 1
	MessageTypeImage = 2
	MessageTypeVoice = 3
	MessageTypeVideo = 4
)

const (
	OfflineReadIndexExpiresIn = time.Hour * 24 * 30 // 读索引在缓存中的过期时间
	OfflineSyncIndexCount     = 2000                // 单次同步消息索引的数量
	OfflineMessageExpiresIn   = 15                  // 离线消息过期时间
	MessageMaxCountPerPage    = 200                 // 同步消息内容时每页的最大数据
)
