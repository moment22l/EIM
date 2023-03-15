package handler

import (
	"EIM"
	"EIM/services/server/service"
	"EIM/wire/pkt"
	"EIM/wire/rpc"
)

type GroupHandler struct {
	groupService service.Group
}

func NewGroupHandler(groupService service.Group) *GroupHandler {
	return &GroupHandler{groupService: groupService}
}

func (h *GroupHandler) DoCreate(ctx EIM.Context) {
	var req pkt.GroupCreateReq
	if err := ctx.ReadBody(&req); err != nil {
		_ = ctx.RespWithError(pkt.Status_InvalidPacketBody, err)
		return
	}
	resp, err := h.groupService.Create(ctx.Session().GetApp(), &rpc.CreateGroupReq{
		Name:         req.GetName(),
		Avatar:       req.GetAvatar(),
		Introduction: req.GetIntroduction(),
		Owner:        req.GetOwner(),
		Members:      req.GetMembers(),
	})
	if err != nil {
		_ = ctx.RespWithError(pkt.Status_SystemException, err)
		return
	}

	locs, err := ctx.GetLocations(req.GetMembers()...)
	if err != nil && err != EIM.ErrSessionNil {
		_ = ctx.RespWithError(pkt.Status_SystemException, err)
		return
	}
	if len(locs) > 0 {
		if err = ctx.Dispatch(&pkt.GroupCreateNotify{
			GroupId: resp.GroupId,
			Members: req.GetMembers(),
		}, locs...); err != nil {
			_ = ctx.RespWithError(pkt.Status_SystemException, err)
			return
		}
	}

	_ = ctx.Resp(pkt.Status_Success, &pkt.GroupCreateResp{GroupId: resp.GroupId})
}

func (h *GroupHandler) DoJoin(ctx EIM.Context) {
	var req pkt.GroupJoinReq
	if err := ctx.ReadBody(&req); err != nil {
		_ = ctx.RespWithError(pkt.Status_InvalidPacketBody, err)
		return
	}
	err := h.groupService.Join(ctx.Session().GetApp(), &rpc.JoinGroupReq{
		Account: req.GetAccount(),
		GroupId: req.GetGroupId(),
	})
	if err != nil {
		_ = ctx.RespWithError(pkt.Status_SystemException, err)
		return
	}

	_ = ctx.Resp(pkt.Status_Success, nil)
}

func (h *GroupHandler) DoQuit(ctx EIM.Context) {
	var req pkt.GroupQuitReq
	if err := ctx.ReadBody(&req); err != nil {
		_ = ctx.RespWithError(pkt.Status_InvalidPacketBody, err)
		return
	}
	err := h.groupService.Quit(ctx.Session().GetApp(), &rpc.QuitGroupReq{
		Account: req.GetAccount(),
		GroupId: req.GetGroupId(),
	})
	if err != nil {
		_ = ctx.RespWithError(pkt.Status_SystemException, err)
		return
	}

	_ = ctx.Resp(pkt.Status_Success, nil)
}

func (h *GroupHandler) DoDetail(ctx EIM.Context) {
	var req pkt.GroupGetReq
	if err := ctx.ReadBody(&req); err != nil {
		_ = ctx.RespWithError(pkt.Status_InvalidPacketBody, err)
		return
	}
	resp, err := h.groupService.Detail(ctx.Session().GetApp(), &rpc.GetGroupReq{
		GroupId: req.GetGroupId(),
	})
	if err != nil {
		_ = ctx.RespWithError(pkt.Status_SystemException, err)
		return
	}

	membersResp, err := h.groupService.Members(ctx.Session().GetApp(), &rpc.GroupMembersReq{GroupId: req.GetGroupId()})
	if err != nil {
		_ = ctx.RespWithError(pkt.Status_SystemException, err)
		return
	}
	var members = make([]*pkt.Member, len(membersResp.GetUsers()))
	for i, m := range membersResp.GetUsers() {
		members[i] = &pkt.Member{
			Account:  m.GetAccount(),
			Alias:    m.GetAlias(),
			Avatar:   m.GetAvatar(),
			JoinTime: m.GetJoinTime(),
		}
	}

	_ = ctx.Resp(pkt.Status_Success, &pkt.GroupGetResp{
		Id:           resp.GetId(),
		Name:         resp.GetName(),
		Avatar:       resp.GetAvatar(),
		Introduction: resp.GetIntroduction(),
		Owner:        resp.GetOwner(),
		Members:      members,
	})
}
