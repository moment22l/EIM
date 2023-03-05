package conf

import (
	"EIM/logger"
	"fmt"

	"github.com/kelseyhightower/envconfig"
	"github.com/spf13/viper"
)

type Config struct {
	ServiceID     string   `envconfig:"serviceId"`
	ServiceName   string   `envconfig:"serviceName"`
	Namespace     string   `envconfig:"namespace"`
	Listen        string   `envconfig:"listen"`
	PublicAddress string   `envconfig:"publicAddress"`
	PublicPort    int      `envconfig:"publicPort"`
	Tags          []string `envconfig:"tags"`
	ConsulURL     string   `envconfig:"consulURL"`
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
	logger.Info(config)

	return &config, nil
}
