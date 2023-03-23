package conf

import (
	"EIM"
	"EIM/logger"
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/kataras/iris/v12/middleware/accesslog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/spf13/viper"
)

type Config struct {
	ServiceID     string
	NodeID        int64
	Listen        string `default:":8080"`
	PublicAddress string
	PublicPort    int `default:"8080"`
	Tags          []string
	ConsulURL     string
	RedisAddrs    string
	Driver        string `default:"mysql"`
	BaseDB        string
	MessageDB     string
	LogLevel      string `default:"INFO"`
}

func (c Config) String() string {
	bts, _ := json.Marshal(c)
	return string(bts)
}

// Init 初始化Config
func Init(file string) (*Config, error) {
	viper.SetConfigFile(file)
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/conf")

	var config Config
	// 读配置文件
	if err := viper.ReadInConfig(); err != nil {
		logger.Warn(err)
	} else {
		// 将配置文件中的数据写入config
		if err := viper.Unmarshal(&config); err != nil {
			return nil, err
		}
	}
	// 填充环境变量
	err := envconfig.Process("EIM", &config)
	if err != nil {
		return nil, err
	}
	// 处理ServiceID为空的情况
	if config.ServiceID == "" {
		localIP := EIM.GetLocalIP()
		config.ServiceID = fmt.Sprintf("royal_%s", strings.ReplaceAll(localIP, ".", ""))
		arr := strings.Split(localIP, ".")
		if len(arr) == 4 {
			suffix, _ := strconv.Atoi(arr[3])
			config.NodeID = int64(suffix)
		}
	}
	// 处理PublicAddress为空的情况
	if config.PublicAddress == "" {
		config.PublicAddress = EIM.GetLocalIP()
	}
	logger.Info(config)
	return &config, nil
}

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
	return redisDB, nil
}

func MakeAccessLog() *accesslog.AccessLog {
	// Initialize a new access log middleware.
	ac := accesslog.File("./access.log")
	// Remove this line to disable logging to console:
	ac.AddOutput(os.Stdout)

	// The default configuration:
	ac.Delim = '|'
	ac.TimeFormat = "2006-01-02 15:04:05"
	ac.Async = false
	ac.IP = true
	ac.BytesReceivedBody = true
	ac.BytesSentBody = true
	ac.BytesReceived = false
	ac.BytesSent = false
	ac.BodyMinify = true
	ac.RequestBody = true
	ac.ResponseBody = false
	ac.KeepMultiLineError = true
	ac.PanicLog = accesslog.LogHandler

	// Default line format if formatter is missing:
	// Time|Latency|Code|Method|Path|IP|Path Params Query Fields|Bytes Received|Bytes Sent|Request|Response|
	return ac
}
