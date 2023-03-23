package handler

import (
	"EIM/services/service/database"
	"EIM/wire/rpc"
	"errors"

	"github.com/kataras/iris/v12"
	"gorm.io/gorm"
)

func (h *ServiceHandler) GroupCreate(c iris.Context) {
	app := c.Params().Get("app")
	var req rpc.CreateGroupReq
	if err := c.ReadBody(&req); err != nil {
		c.StopWithError(iris.StatusBadRequest, err)
		return
	}
	groupID := h.IDGen.Next()
	g := &database.Group{
		Model: database.Model{
			ID: groupID.Int64(),
		},
		Group:        groupID.Base36(),
		App:          app,
		Name:         req.Name,
		Owner:        req.Owner,
		Avatar:       req.Avatar,
		Introduction: req.Introduction,
	}
	members := make([]database.GroupMember, len(req.Members))
	for i, m := range req.Members {
		members[i] = database.GroupMember{
			Model: database.Model{
				ID: h.IDGen.Next().Int64(),
			},
			Account: m,
			Group:   groupID.Base36(),
		}
	}
	err := h.BaseDB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(g).Error; err != nil {
			return err
		}
		if err := tx.Create(&members).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		c.StopWithError(iris.StatusInternalServerError, err)
		return
	}
	_, _ = c.Negotiate(&rpc.CreateGroupResp{GroupId: groupID.Base36()})
}

func (h *ServiceHandler) GroupJoin(c iris.Context) {
	var req rpc.JoinGroupReq
	if err := c.ReadBody(&req); err != nil {
		c.StopWithError(iris.StatusBadRequest, err)
		return
	}
	gm := &database.GroupMember{
		Model: database.Model{
			ID: h.IDGen.Next().Int64(),
		},
		Account: req.Account,
		Group:   req.GroupId,
	}
	err := h.BaseDB.Create(gm).Error
	if err != nil {
		c.StopWithError(iris.StatusInternalServerError, err)
		return
	}
}

func (h *ServiceHandler) GroupQuit(c iris.Context) {
	var req rpc.QuitGroupReq
	if err := c.ReadBody(&req); err != nil {
		c.StopWithError(iris.StatusBadRequest, err)
		return
	}
	gm := &database.GroupMember{
		Account: req.Account,
		Group:   req.GroupId,
	}
	err := h.BaseDB.Delete(&database.GroupMember{}, gm).Error
	if err != nil {
		c.StopWithError(iris.StatusInternalServerError, err)
		return
	}
}

func (h *ServiceHandler) GroupMembers(c iris.Context) {
	group := c.Params().Get("id")
	if group == "" {
		c.StopWithError(iris.StatusBadRequest, errors.New("group is null"))
		return
	}
	var members []database.GroupMember
	err := h.BaseDB.Order("Updated_At asc").Find(&members, database.GroupMember{Group: group}).Error
	if err != nil {
		c.StopWithError(iris.StatusInternalServerError, err)
		return
	}
	var users = make([]*rpc.Member, len(members))
	for i, m := range members {
		users[i] = &rpc.Member{
			Account:  m.Account,
			Alias:    m.Alias,
			JoinTime: m.CreatedAt.Unix(),
		}
	}
	_, _ = c.Negotiate(&rpc.GroupMembersResp{
		Users: users,
	})
}

func (h *ServiceHandler) GroupGet(c iris.Context) {
	groupId := c.Params().Get("id")
	if groupId == "" {
		c.StopWithError(iris.StatusBadRequest, errors.New("groupId is null"))
		return
	}
	id, err := h.IDGen.ParseBase36(groupId)
	if err != nil {
		c.StopWithError(iris.StatusBadRequest, errors.New("group is invalid"+groupId))
		return
	}
	var g database.Group
	err = h.BaseDB.First(&g, id.Int64()).Error
	if err != nil {
		c.StopWithError(iris.StatusInternalServerError, err)
		return
	}
	_, _ = c.Negotiate(&rpc.GetGroupResp{
		Id:           id.Base36(),
		Name:         g.Name,
		Avatar:       g.Avatar,
		Introduction: g.Introduction,
		Owner:        g.Owner,
		CreatedAt:    g.CreatedAt.Unix(),
	})
}
