package handler

import (
	"EIM"
	"EIM/logger"
	"EIM/wire/pkt"
)

// LoginHandler 登录管理
type LoginHandler struct {
}

func NewLoginHandler() *LoginHandler {
	return &LoginHandler{}
}

// DoSysLogin 登录
func (h *LoginHandler) DoSysLogin(ctx EIM.Context) {
	// 序列化
	var session pkt.Session
	err := ctx.ReadBody(&session)
	if err != nil {
		_ = ctx.RespWithError(pkt.Status_InvalidPacketBody, err)
		return
	}
	logger.WithFields(logger.Fields{
		"Func":      "Login",
		"ChannelId": session.GetChannelId(),
		"Account":   session.GetAccount(),
		"RemoteIP":  session.GetRemoteIP(),
	}).Info("do login")
	// 检测该账号是否已登录
	old, err := ctx.GetLocation(session.Account, "")
	if err != nil && err != EIM.ErrSessionNil {
		_ = ctx.RespWithError(pkt.Status_SystemException, err)
		return
	}
	// 通知老用户下线
	if old != nil {
		_ = ctx.Dispatch(&pkt.KickoutNotify{
			ChannelId: old.ChannelId,
		}, old)
	}
	// 将新用户添加到会话管理器中
	err = ctx.Add(&session)
	if err != nil {
		_ = ctx.RespWithError(pkt.Status_SystemException, err)
		return
	}
	// 通知登录成功
	var resp = &pkt.LoginResp{
		ChannelId: session.ChannelId,
	}
	_ = ctx.Resp(pkt.Status_Success, resp)
}

// DoSysLogout 登出
func (h *LoginHandler) DoSysLogout(ctx EIM.Context) {
	logger.WithFields(logger.Fields{
		"Func":      "Logout",
		"ChannelId": ctx.Session().GetChannelId(),
		"Account":   ctx.Session().GetAccount(),
	}).Info("do Logout ")

	err := ctx.Delete(ctx.Session().GetAccount(), ctx.Session().GetChannelId())
	if err != nil {
		_ = ctx.RespWithError(pkt.Status_SystemException, err)
		return
	}

	_ = ctx.Resp(pkt.Status_Success, nil)
}
