package datebase

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// KeyMessageAckIndex 返回读索引的一个redis key
func KeyMessageAckIndex(account string) string {
	return fmt.Sprintf("chat:ack:%s", account)
}

// InitRedis 返回一个redis实例
func InitRedis(addr string, password string) (*redis.Client, error) {
	redisDB := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DialTimeout:  time.Second * 5,
		ReadTimeout:  time.Second * 5,
		WriteTimeout: time.Second * 5,
	})

	_, err := redisDB.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}
	return redisDB, err
}
