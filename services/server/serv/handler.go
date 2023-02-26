package serv

import (
	"EIM"
	"EIM/container"
	"EIM/logger"
	"EIM/wire"
	"EIM/wire/pkt"
	"strings"
)

var log = logger.WithFields(logger.Fields{
	"module": wire.SNChat,
	"pkg":    "serv",
})

type ServHandler struct {
}

func (h *ServHandler) Receive(ag EIM.Agent, payload []byte) {
	//TODO implement me
	panic("implement me")
}

type ServerDispatcher struct{}

func (s *ServerDispatcher) Push(gateway string, channels []string, pkt *pkt.LoginPkt) error {
	pkt.AddStringMeta(wire.MetaDestChannels, strings.Join(channels, ","))
	return container.Push(gateway, pkt)
}
