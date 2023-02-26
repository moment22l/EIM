package EIM

import (
	"EIM/wire/pkt"
	"errors"
)

var ErrSessionNil = errors.New("err:session nil")

// SessionStorage 会话管理器, 提供保存、删除、查询会话等功能
type SessionStorage interface {
	Add(session *pkt.Session) error
	Delete(account string, channelId string) error
	Get(channelId string) (*pkt.Session, error)
	GetLocations(accounts ...string) ([]*Location, error)
	GetLocation(account string, device string) (*Location, error)
}
