package EIM

import (
	"EIM/logger"
	"sync"
)

// ChannelMap 连接管理
type ChannelMap interface {
	Add(channel Channel)
	Remove(id string)
	Get(id string) (channel Channel, ok bool)
	All() []Channel
}

// ChannelsImp channels实现
type ChannelsImp struct {
	channels *sync.Map
}

// NewChannels 创建ChannelsImp
func NewChannels(num int) *ChannelsImp {
	return &ChannelsImp{
		channels: new(sync.Map),
	}
}

// Add 向c中添加channel
func (ch *ChannelsImp) Add(channel Channel) {
	if channel.ID() == "" {
		logger.WithFields(logger.Fields{
			"module": "ChannelsImpl",
		}).Error("channel id is required")
	}

	ch.channels.Store(channel.ID(), channel)
}

// Remove 从c中移除对应id的channel
func (ch *ChannelsImp) Remove(id string) {
	ch.channels.Delete(id)
}

// Get 从c中获取对应id的channel
func (ch *ChannelsImp) Get(id string) (Channel, bool) {
	if id == "" {
		logger.WithFields(logger.Fields{
			"module": "ChannelsImpl",
		}).Error("channel id is required")
	}
	val, ok := ch.channels.Load(id)
	if !ok {
		return nil, false
	}
	return val.(Channel), true
}

// All 返回c中所有channel
func (ch *ChannelsImp) All() []Channel {
	arr := make([]Channel, 0)
	ch.channels.Range(func(key, val any) bool {
		arr = append(arr, val.(Channel))
		return true
	})
	return arr
}
