package db

import (
	"12305/config"
	"12305/model"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	DB    *gorm.DB
	Redis *redis.Client
)

func initViper() {
	if err := config.Init(""); err != nil {
		panic("failed to init config")
	}
}

func InitDatabase() {
	conf := &model.DBConf{
		Host:     viper.GetString("database.host"),
		Port:     viper.GetString("database.port"),
		User:     viper.GetString("database.username"),
		Password: viper.GetString("database.password"),
		DBName:   viper.GetString("database.name"),
	}

	config := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		conf.User,
		conf.Password,
		conf.Host,
		conf.Port,
		conf.DBName,
	)
	var err error
	DB, err = gorm.Open(mysql.Open(config), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
}

func InitRedis() {
	conf := &model.RedisConf{
		Host:     viper.GetString("redis.host"),
		Port:     viper.GetString("redis.port"),
		Password: viper.GetString("redis.password"),
	}
	Redis = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", conf.Host, conf.Port),
		Password: conf.Password,
	})
}

func init() {
	initViper()
	InitDatabase()
	InitRedis()
}
