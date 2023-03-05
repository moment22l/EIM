package storage

import (
	"EIM"
	"EIM/wire/pkt"
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"google.golang.org/protobuf/proto"
)

const (
	LocationExpired = time.Hour * 48
)

type RedisStorage struct {
	cli *redis.Client
}

func NewRedisStorage(cli *redis.Client) EIM.SessionStorage {
	return &RedisStorage{
		cli: cli,
	}
}

// Add 将会话添加进redis缓存
func (r *RedisStorage) Add(session *pkt.Session) error {
	ctx := context.Background()
	// 保存location
	loc := &EIM.Location{
		ChannelId: session.ChannelId,
		GateId:    session.GateId,
	}
	locKey := KeyLocation(session.Account, "")
	err := r.cli.Set(ctx, locKey, loc.Bytes(), LocationExpired).Err()
	if err != nil {
		return err
	}
	// 保存Session
	snKey := KeySession(session.Account)
	buf, _ := proto.Marshal(session)
	err = r.cli.Set(ctx, snKey, buf, LocationExpired).Err()
	if err != nil {
		return err
	}
	return nil
}

// Delete 将会话从redis缓存中删除
func (r *RedisStorage) Delete(account string, channelId string) error {
	ctx := context.Background()
	// 删除location
	locKey := KeyLocation(account, "")
	err := r.cli.Del(ctx, locKey).Err()
	if err != nil {
		return err
	}
	// 删除Session
	snKey := KeySession(account)
	err = r.cli.Del(ctx, snKey).Err()
	if err != nil {
		return err
	}
	return nil
}

// Get 从redis中获取channelId对应的会话
func (r *RedisStorage) Get(channelId string) (*pkt.Session, error) {
	ctx := context.Background()
	snKey := KeySession(channelId)
	bs, err := r.cli.Get(ctx, snKey).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, EIM.ErrSessionNil
		}
		return nil, err
	}
	var session *pkt.Session
	_ = proto.Unmarshal(bs, session)
	return session, nil
}

// GetLocations 获取多个用户的位置
func (r *RedisStorage) GetLocations(accounts ...string) ([]*EIM.Location, error) {
	// TODO：可采用 r.cli.MGet 进行性能优化, 一次性获取结果, 减少批量Get操作时导致的网络来回耗时
	locs := make([]*EIM.Location, 0)
	for _, account := range accounts {
		loc, _ := r.GetLocation(account, "")
		if loc == nil {
			continue
		}
		locs = append(locs, loc)
	}
	if len(locs) == 0 {
		return nil, EIM.ErrSessionNil
	}
	return locs, nil
}

// GetLocation 获取单个用户的位置
func (r *RedisStorage) GetLocation(account string, device string) (*EIM.Location, error) {
	ctx := context.Background()
	locKey := KeyLocation(account, device)
	bs, err := r.cli.Get(ctx, locKey).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, EIM.ErrSessionNil
		}
		return nil, err
	}
	var loc *EIM.Location
	_ = loc.Unmarshal(bs)
	return loc, nil
}

// KeySession 根据channel生成session的key
func KeySession(channel string) string {
	return fmt.Sprintf("login:sn:%s", channel)
}

// KeyLocation 生成一个location的key
func KeyLocation(account, device string) string {
	if device == "" {
		return fmt.Sprintf("login:loc:%s", account)
	}
	return fmt.Sprintf("login:loc:%s:%s", account, device)
}

// KeyLocations 生成多个location的key
func KeyLocations(accounts ...string) []string {
	arr := make([]string, len(accounts))
	for i, account := range accounts {
		arr[i] = KeyLocation(account, "")
	}
	return arr
}
