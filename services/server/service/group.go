package service

import (
	"EIM/logger"
	"EIM/wire/rpc"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	"google.golang.org/protobuf/proto"
)

type Group interface {
	Create(app string, req *rpc.CreateGroupReq) (*rpc.CreateGroupResp, error)
	Members(app string, req *rpc.GroupMembersReq) (*rpc.GroupMembersResp, error)
	Join(app string, req *rpc.JoinGroupReq) error
	Quit(app string, req *rpc.QuitGroupReq) error
	Detail(app string, req *rpc.GetGroupReq) (*rpc.GetGroupResp, error)
}

type GroupHttp struct {
	url string
	cli *resty.Client
	srv *resty.SRVRecord
}

// Create 创建组
func (g *GroupHttp) Create(app string, req *rpc.CreateGroupReq) (*rpc.CreateGroupResp, error) {
	path := fmt.Sprintf("%s/api/%s/group", g.url, app)
	body, _ := proto.Marshal(req)
	response, err := g.Req().SetBody(body).Post(path)
	if err != nil {
		return nil, err
	}
	if response.StatusCode() != 200 {
		return nil, fmt.Errorf("GroupHttp.Create response.StatusCode() = %d, but want 200", response.StatusCode())
	}
	var resp rpc.CreateGroupResp
	_ = proto.Unmarshal(response.Body(), &resp)
	logger.Debugf("GroupHttp.Create resp: %v", &resp)
	return &resp, nil
}

// Members 返回所有组成员
func (g *GroupHttp) Members(app string, req *rpc.GroupMembersReq) (*rpc.GroupMembersResp, error) {
	path := fmt.Sprintf("%s/api/%s/group/members/%s", g.url, app, req.GroupId)
	response, err := g.Req().Get(path)
	if err != nil {
		return nil, err
	}
	if response.StatusCode() != 200 {
		return nil, fmt.Errorf("GroupHttp.Members response.StatusCode() = %d, but want 200", response.StatusCode())
	}
	var resp rpc.GroupMembersResp
	_ = proto.Unmarshal(response.Body(), &resp)
	logger.Errorf("GroupHttp.Members response: %v", &resp)
	return &resp, nil
}

// Join 加入组
func (g *GroupHttp) Join(app string, req *rpc.JoinGroupReq) error {
	path := fmt.Sprintf("%s/api/%s/group/member", g.url, app)
	body, _ := proto.Marshal(req)
	response, err := g.Req().SetBody(body).Post(path)
	if err != nil {
		return err
	}
	if response.StatusCode() != 200 {
		return fmt.Errorf("GroupHttp.Join response.StatusCode() = %d, but want 200", response.StatusCode())
	}
	return nil
}

// Quit 退出组
func (g *GroupHttp) Quit(app string, req *rpc.QuitGroupReq) error {
	path := fmt.Sprintf("%s/api/%s/group/member", g.url, app)
	body, _ := proto.Marshal(req)
	response, err := g.Req().SetBody(body).Delete(path)
	if err != nil {
		return err
	}
	if response.StatusCode() != 200 {
		return fmt.Errorf("GroupHttp.Quit response.StatusCode() = %d, but want 200", response.StatusCode())
	}
	return nil
}

// Detail 组信息
func (g *GroupHttp) Detail(app string, req *rpc.GetGroupReq) (*rpc.GetGroupResp, error) {
	path := fmt.Sprintf("%s/api/%s/group/%s", g.url, app, req.GroupId)
	response, err := g.Req().Get(path)
	if err != nil {
		return nil, err
	}
	if response.StatusCode() != 200 {
		return nil, fmt.Errorf("GroupHttp.Detial response.StatusCode() = %d, but want 200", response.StatusCode())
	}
	var resp rpc.GetGroupResp
	_ = proto.Unmarshal(response.Body(), &resp)
	logger.Debugf("GroupHttp.Detail resp: %v", &resp)
	return &resp, nil
}

func (g *GroupHttp) Req() *resty.Request {
	if g.srv == nil {
		return g.cli.R()
	}
	return g.cli.R().SetSRV(g.srv)
}

func NewGroupService(url string) Group {
	cli := resty.New().SetRetryCount(5).SetTimeout(time.Second * 5)
	cli.SetHeader("Content-type", "application/x-protobuf")
	cli.SetHeader("Accept", "application/x-protobuf")
	cli.SetScheme("http")

	return &GroupHttp{
		url: url,
		cli: cli,
	}
}

func NewGroupServiceWithSRV(scheme string, srv *resty.SRVRecord) Group {
	cli := resty.New().SetRetryCount(5).SetTimeout(time.Second * 5)
	cli.SetHeader("Content-type", "application/x-protobuf")
	cli.SetHeader("Accept", "application/x-protobuf")
	cli.SetScheme(scheme)

	return &GroupHttp{
		url: "",
		cli: cli,
		srv: srv,
	}
}
