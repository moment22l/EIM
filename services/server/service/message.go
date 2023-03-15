package service

import (
	"EIM/logger"
	"EIM/wire/rpc"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	"google.golang.org/protobuf/proto"
)

type Message interface {
	InsertUser(app string, req *rpc.InsertMessageReq) (*rpc.InsertMessageResp, error)
	InsertGroup(app string, req *rpc.InsertMessageReq) (*rpc.InsertMessageResp, error)
	SetAck(app string, req *rpc.AckMessageReq) error
	GetMessageIndex(app string, req *rpc.GetOfflineMessageIndexReq) (*rpc.GetOfflineMessageIndexResp, error)
	GetMessageContent(app string, req *rpc.GetOfflineMessageContentReq) (*rpc.GetOfflineMessageContentResp, error)
}

type MessageHttp struct {
	url string
	cli *resty.Client
	srv *resty.SRVRecord
}

// InsertUser 插入用户
func (m *MessageHttp) InsertUser(app string, req *rpc.InsertMessageReq) (*rpc.InsertMessageResp, error) {
	path := fmt.Sprintf("%s/api/%s/message/user", m.url, app)
	t := time.Now()
	body, _ := proto.Marshal(req)
	response, err := m.Req().SetBody(body).Post(path)
	if err != nil {
		return nil, err
	}
	if response.StatusCode() != 200 {
		return nil, fmt.Errorf("MessageHttp.InsertUser response.StatusCode() = %d, but want 200", response.StatusCode())
	}
	var resp rpc.InsertMessageResp
	_ = proto.Unmarshal(response.Body(), &resp)
	logger.Debugf("messageHttp.InsertUser cost %v, resp: %v", time.Since(t), &resp)
	return &resp, nil
}

// InsertGroup 插入组
func (m *MessageHttp) InsertGroup(app string, req *rpc.InsertMessageReq) (*rpc.InsertMessageResp, error) {
	path := fmt.Sprintf("%s/api/%s/message/group", m.url, app)
	t := time.Now()
	body, _ := proto.Marshal(req)
	response, err := m.Req().SetBody(body).Post(path)
	if err != nil {
		return nil, err
	}
	if response.StatusCode() != 200 {
		return nil, fmt.Errorf("MessageHttp.InsertGroup response.StatusCode() = %d, but want 200", response.StatusCode())
	}
	var resp rpc.InsertMessageResp
	_ = proto.Unmarshal(response.Body(), &resp)
	logger.Debugf("messageHttp.InsertGroup cost %v, resp: %v", time.Since(t), &resp)
	return &resp, nil
}

// SetAck 设置Ack
func (m *MessageHttp) SetAck(app string, req *rpc.AckMessageReq) error {
	path := fmt.Sprintf("%s/api/%s/message/ack", m.url, app)
	body, _ := proto.Marshal(req)
	response, err := m.Req().SetBody(body).Post(path)
	if err != nil {
		return err
	}
	if response.StatusCode() != 200 {
		return fmt.Errorf("MessageHttp.SetAck response.StatusCode() = %d, but want 200", response.StatusCode())
	}
	return nil
}

// GetMessageIndex 获取离线消息索引
func (m *MessageHttp) GetMessageIndex(app string, req *rpc.GetOfflineMessageIndexReq) (*rpc.GetOfflineMessageIndexResp, error) {
	path := fmt.Sprintf("%s/api/%s/offline/index", m.url, app)
	t := time.Now()
	body, _ := proto.Marshal(req)
	response, err := m.Req().SetBody(body).Post(path)
	if err != nil {
		return nil, err
	}
	if response.StatusCode() != 200 {
		return nil, fmt.Errorf("MessageHttp.GetMessageIndex response.StatusCode() = %d, but want 200", response.StatusCode())
	}
	var resp rpc.GetOfflineMessageIndexResp
	_ = proto.Unmarshal(response.Body(), &resp)
	logger.Debugf("messageHttp.GetMessageIndex cost %v, resp: %v", time.Since(t), &resp)
	return &resp, nil
}

// GetMessageContent 获取离线消息内容
func (m *MessageHttp) GetMessageContent(app string, req *rpc.GetOfflineMessageContentReq) (*rpc.GetOfflineMessageContentResp, error) {
	path := fmt.Sprintf("%s/api/%s/offline/content", m.url, app)
	t := time.Now()
	body, _ := proto.Marshal(req)
	response, err := m.Req().SetBody(body).Post(path)
	if err != nil {
		return nil, err
	}
	if response.StatusCode() != 200 {
		return nil, fmt.Errorf("MessageHttp.GetMessageContent response.StatusCode() = %d, but want 200", response.StatusCode())
	}
	var resp rpc.GetOfflineMessageContentResp
	_ = proto.Unmarshal(response.Body(), &resp)
	logger.Debugf("messageHttp.GetMessageContent cost %v, resp: %v", time.Since(t), &resp)
	return &resp, nil
}

func (m *MessageHttp) Req() *resty.Request {
	if m.srv == nil {
		return m.cli.R()
	}
	return m.cli.R().SetSRV(m.srv)
}

func NewMessageService(url string) Message {
	cli := resty.New().SetRetryCount(3).SetTimeout(time.Second * 5)
	cli.SetHeader("Content-type", "application/x-protobuf")
	cli.SetHeader("Accept", "application/x-protobuf")
	cli.SetScheme("http")
	return &MessageHttp{
		url: url,
		cli: cli,
	}
}

func NewMessageServiceWithSRV(scheme string, srv *resty.SRVRecord) Message {
	cli := resty.New().SetRetryCount(3).SetTimeout(time.Second * 5)
	cli.SetHeader("Content-type", "application/x-protobuf")
	cli.SetHeader("Accept", "application/x-protobuf")
	cli.SetScheme(scheme)
	return &MessageHttp{
		url: "",
		cli: cli,
		srv: srv,
	}
}
