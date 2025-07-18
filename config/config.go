package config

import (
	"fmt"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

const PageSize = 10

type Config struct {
	Name string
}

func Init(name string) error {
	c := Config{
		Name: name,
	}
	if err := c.InitConfig(); err != nil {
		return err
	}

	c.WatchConfig()
	return nil
}

// viper库初始化config
func (c *Config) InitConfig() error {
	if c.Name != "" {
		viper.SetConfigName(c.Name)
	} else {
		viper.AddConfigPath("conf")
		viper.SetConfigName("config")
	}
	viper.SetConfigType("yaml")
	err := viper.ReadInConfig()
	if err != nil {
		return err
	}
	return nil
}

// 监控配置文件变化并热加载
func (c *Config) WatchConfig() {
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("Config file changed:", e.Name)
	})
}
