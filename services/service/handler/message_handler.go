package handler

import (
	"EIM/services/service/database"
	"EIM/wire"
	"EIM/wire/rpc"
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/kataras/iris/v12"
	"gorm.io/gorm"
	"time"
)

type ServiceHandler struct {
	BaseDB    *gorm.DB
	MessageDB *gorm.DB
	Cache     *redis.Client
	IDGen     *database.IDGenerator
}

// InsertUserMessage 插入单聊消息对应的数据到数据库中
func (h *ServiceHandler) InsertUserMessage(ctx iris.Context) {
	var req rpc.InsertMessageReq
	if err := ctx.ReadBody(&req); err != nil {
		ctx.StopWithError(iris.StatusBadRequest, err)
		return
	}
	messageID := h.IDGen.Next().Int64()
	// 消息内容
	content := database.MessageContent{
		ID:       messageID,
		Type:     byte(req.GetMessage().GetType()),
		Body:     req.GetMessage().GetBody(),
		Extra:    req.GetMessage().GetExtra(),
		SendTime: req.GetSendTime(),
	}
	// 消息索引
	ids := make([]database.MessageIndex, 2)
	ids[0] = database.MessageIndex{
		ID:        h.IDGen.Next().Int64(),
		AccountA:  req.GetSender(),
		AccountB:  req.GetDest(),
		Direction: 1,
		MessageID: messageID,
		SendTime:  req.GetSendTime(),
	}
	ids[1] = database.MessageIndex{
		ID:        h.IDGen.Next().Int64(),
		AccountA:  req.GetDest(),
		AccountB:  req.GetSender(),
		Direction: 0,
		MessageID: messageID,
		SendTime:  req.GetSendTime(),
	}
	// 创建事务将数据存入数据库
	err := h.MessageDB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&content).Error; err != nil {
			return err
		}
		if err := tx.Create(&ids).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		ctx.StopWithError(iris.StatusInternalServerError, err)
		return
	}
	_, _ = ctx.Negotiate(&rpc.InsertMessageResp{MessageId: messageID})
}

// InsertGroupMessage 插入群聊消息对应的数据到数据库中
func (h *ServiceHandler) InsertGroupMessage(ctx iris.Context) {
	var req rpc.InsertMessageReq
	if err := ctx.ReadBody(&req); err != nil {
		ctx.StopWithError(iris.StatusBadRequest, err)
		return
	}
	messageID := h.IDGen.Next().Int64()
	// 消息内容
	content := database.MessageContent{
		ID:       messageID,
		Type:     byte(req.GetMessage().GetType()),
		Body:     req.GetMessage().GetBody(),
		Extra:    req.GetMessage().GetExtra(),
		SendTime: req.GetSendTime(),
	}
	// 找到群中的所有用户
	var members []database.GroupMember
	err := h.BaseDB.Where(&database.GroupMember{Group: req.Dest}).Find(&members).Error
	if err != nil {
		ctx.StopWithError(iris.StatusInternalServerError, err)
		return
	}
	// 消息索引
	ids := make([]database.MessageIndex, len(members))
	for i, m := range members {
		ids[i] = database.MessageIndex{
			ID:        h.IDGen.Next().Int64(),
			AccountA:  m.Account,
			AccountB:  req.GetSender(),
			Direction: 0,
			MessageID: messageID,
			Group:     req.GetDest(),
			SendTime:  req.GetSendTime(),
		}
		if ids[i].AccountA == ids[i].AccountB {
			ids[i].Direction = 1
		}
	}
	// 创建事务将数据存入数据库
	err = h.MessageDB.Transaction(func(tx *gorm.DB) error {
		if err = tx.Create(&content).Error; err != nil {
			return err
		}
		if err = tx.Create(&ids).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		ctx.StopWithError(iris.StatusInternalServerError, err)
		return
	}
	_, _ = ctx.Negotiate(&rpc.InsertMessageResp{MessageId: messageID})
}

// MessageAck 根据Ack包重置读索引
func (h *ServiceHandler) MessageAck(ctx iris.Context) {
	var req rpc.AckMessageReq
	if err := ctx.ReadBody(&req); err != nil {
		ctx.StopWithError(iris.StatusBadRequest, err)
		return
	}
	// 保存到redis中
	err := setMessageAck(h.Cache, req.Account, req.MessageId)
	if err != nil {
		ctx.StopWithError(iris.StatusInternalServerError, err)
		return
	}
}

// setMessageAck 重置读索引
func setMessageAck(cache *redis.Client, account string, msgID int64) error {
	if msgID == 0 {
		return nil
	}
	key := database.KeyMessageAckIndex(account)
	return cache.Set(context.Background(), key, msgID, wire.OfflineReadIndexExpiresIn).Err()
}

// GetOfflineMessageIndex 同步离线消息索引
func (h *ServiceHandler) GetOfflineMessageIndex(ctx iris.Context) {
	var req rpc.GetOfflineMessageIndexReq
	if err := ctx.ReadBody(&req); err != nil {
		ctx.StopWithError(iris.StatusBadRequest, err)
		return
	}
	msgID := req.GetMessageId()
	// 获取读索引全局时钟
	start, err := h.getSentTime(req.Account, req.MessageId)
	if err != nil {
		ctx.StopWithError(iris.StatusInternalServerError, err)
		return
	}
	// 从DB中加载消息索引列表
	var indexes []*rpc.MessageIndex
	tx := h.MessageDB.Model(&database.MessageIndex{}).Select("send_time", "account_b", "direction", "group")
	err = tx.Where("account_a=? and send_time>? and direction=?", req.GetAccount(), start, 0).
		Order("send_time asc").Limit(wire.OfflineSyncIndexCount).Find(indexes).Error
	if err != nil {
		ctx.StopWithError(iris.StatusInternalServerError, err)
		return
	}
	// 重置读索引
	err = setMessageAck(h.Cache, req.GetAccount(), msgID)
	if err != nil {
		ctx.StopWithError(iris.StatusInternalServerError, err)
		return
	}
	_, _ = ctx.Negotiate(&rpc.GetOfflineMessageIndexResp{List: indexes})
}

// getSentTime 获取全局时钟
func (h *ServiceHandler) getSentTime(account string, msgID int64) (int64, error) {
	// 冷启动(指在某个设备中第一次启动。或者在web端没有本地消息存储的情况下，第一次同步索引，消息ID就会为空)
	if msgID == 0 {
		key := database.KeyMessageAckIndex(account)
		msgID, _ = h.Cache.Get(context.Background(), key).Int64()
	}
	var start int64
	if msgID > 0 {
		var content database.MessageContent
		err := h.MessageDB.Select("send_time").First(&content, &msgID).Error
		if err != nil {
			start = time.Now().AddDate(0, 0, -1).UnixNano()
		} else {
			start = content.SendTime
		}
	}
	// 返回默认的离线消息过期时间
	earliestKeepTime := time.Now().AddDate(0, 0, -1*wire.OfflineMessageExpiresIn).UnixNano()
	if start == 0 || start < earliestKeepTime {
		start = earliestKeepTime
	}
	return start, nil
}

// GetOfflineMessageContent 同步离线消息内容
func (h *ServiceHandler) GetOfflineMessageContent(ctx iris.Context) {
	var req rpc.GetOfflineMessageContentReq
	if err := ctx.ReadBody(&req); err != nil {
		ctx.StopWithError(iris.StatusBadRequest, err)
		return
	}
	msgIds := len(req.GetMessageIds())
	if msgIds > wire.MessageMaxCountPerPage {
		ctx.StopWithText(iris.StatusInternalServerError, "too much msgIds")
		return
	}

	var contents []*rpc.Message
	err := h.MessageDB.Model(&database.MessageContent{}).Where(msgIds).Find(&contents).Error
	if err != nil {
		ctx.StopWithError(iris.StatusInternalServerError, err)
		return
	}
	_, _ = ctx.Negotiate(&rpc.GetOfflineMessageContentResp{List: contents})
}
