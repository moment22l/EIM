package conf

import (
	"EIM/logger"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/kelseyhightower/envconfig"
	"github.com/spf13/viper"
)

type Config struct {
	ServiceID     string   `envconfig:"serviceId"`
	Namespace     string   `envconfig:"namespace"`
	Listen        string   `envconfig:"listen"`
	PublicAddress string   `envconfig:"publicAddress"`
	PublicPort    int      `envconfig:"publicPort"`
	Tags          []string `envconfig:"tags"`
	ConsulURL     string   `envconfig:"consulURL"`
	RedisAddr     string   `envconfig:"redisAddr"`
	RpcURL        string   `envconfig:"rpcURL"`
}

// Init 初始化配置
func Init(file string) (*Config, error) {
	viper.SetConfigFile(file)
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/conf")

	err := viper.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("config file not found: %w", err)
	}

	var config Config
	err = viper.Unmarshal(&config)
	if err != nil {
		return nil, err
	}
	err = envconfig.Process("", &config)
	if err != nil {
		return nil, err
	}

	logger.Info(err)
	return &config, nil
}

// InitRedis 初始化redis
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
		log.Println(err)
		return nil, err
	}
	return redisDB, nil
}
