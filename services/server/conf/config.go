package conf

import (
	"EIM/logger"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/kelseyhightower/envconfig"
	"github.com/spf13/viper"
)

type Config struct {
	ServiceID       string
	Listen          string `default:":8005"`
	MonitorPort     int    `default:"8006"`
	PublicAddress   string
	PublicPort      int `default:"8005"`
	Tags            []string
	Zone            string `default:"zone_ali_03"`
	ConsulURL       string
	RedisAddrs      string
	RoyalURL        string
	LogLevel        string `default:"DEBUG"`
	MessageGPool    int    `default:"5000"`
	ConnectionGPool int    `default:"500"`
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

func (c Config) String() string {
	bts, _ := json.Marshal(c)
	return string(bts)
}
