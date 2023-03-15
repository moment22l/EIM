package handler

import (
	"EIM"
	"EIM/services/server/service"
	"EIM/wire/pkt"
	"EIM/wire/rpc"
	"errors"
)

type OfflineHandler struct {
	msgService service.Message
}

func NewOfflineHandler(msgService service.Message) *OfflineHandler {
	return &OfflineHandler{msgService: msgService}
}

// DoSyncIndex 同步离线消息索引
func (h *OfflineHandler) DoSyncIndex(ctx EIM.Context) {
	var req pkt.MessageIndexReq
	if err := ctx.ReadBody(&req); err != nil {
		_ = ctx.RespWithError(pkt.Status_InvalidPacketBody, err)
	}
	resp, err := h.msgService.GetMessageIndex(ctx.Session().GetApp(), &rpc.GetOfflineMessageIndexReq{
		Account:   ctx.Session().GetAccount(),
		MessageId: req.GetMessageId(),
	})
	if err != nil {
		_ = ctx.RespWithError(pkt.Status_SystemException, err)
		return
	}
	list := make([]*pkt.MessageIndex, len(resp.GetList()))
	for i, val := range resp.GetList() {
		list[i] = &pkt.MessageIndex{
			MessageId: val.GetMessageId(),
			Direction: val.GetDirection(),
			SendTime:  val.GetSendTime(),
			AccountB:  val.GetAccountB(),
			Group:     val.GetGroup(),
		}
	}
	_ = ctx.Resp(pkt.Status_Success, &pkt.MessageIndexResp{Indexes: list})
}

// DoSyncContent 同步离线消息内容
func (h *OfflineHandler) DoSyncContent(ctx EIM.Context) {
	var req pkt.MessageContentReq
	if err := ctx.ReadBody(&req); err != nil {
		_ = ctx.RespWithError(pkt.Status_InvalidPacketBody, err)
	}
	if len(req.MessageIds) == 0 {
		_ = ctx.RespWithError(pkt.Status_InvalidPacketBody, errors.New("empty MessageIds"))
		return
	}
	resp, err := h.msgService.GetMessageContent(ctx.Session().GetApp(), &rpc.GetOfflineMessageContentReq{
		MessageIds: req.GetMessageIds(),
	})
	if err != nil {
		_ = ctx.RespWithError(pkt.Status_SystemException, err)
		return
	}
	list := make([]*pkt.MessageContent, len(resp.GetList()))
	for i, val := range resp.GetList() {
		list[i] = &pkt.MessageContent{
			MessageId: val.GetId(),
			Type:      val.GetType(),
			Body:      val.GetBody(),
			Extra:     val.GetExtra(),
		}
	}
	_ = ctx.Resp(pkt.Status_Success, &pkt.MessageContentResp{Contents: list})
}
