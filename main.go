package main

import (
	"12305/api"
	"12305/api/handler"
	"12305/config"
	"12305/db"
	"12305/mq/receiver"
	"12305/mq/sender"
	"12305/repository"
	"12305/service"
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/spf13/viper"
)

func initViper() {
	if err := config.Init(""); err != nil {
		panic("failed to init config")
	}
}

var (
	UserHandler   handler.UserHandler
	TicketHandler handler.TicketHandler
	OrderHandler  handler.OrderHandler
)

func initHandler() {
	// 初始化用户处理器
	UserHandler = handler.UserHandler{
		UserService: &service.UserService{
			UserRepo: repository.UserRepository{
				DB: db.DB,
			},
		},
	}

	// 初始化票务处理器
	TicketHandler = handler.TicketHandler{
		TicketService: &service.TicketService{
			TicketRepo: repository.TicketRepository{
				DB:  db.DB,
				Rdb: db.Redis,
			},
			RedisRepo: repository.RedisRepository{
				Rdb: db.Redis,
			},
			LocalRepo: repository.LocalRepository{},
			RabbitmqRepo: sender.SenderStruct{
				Conn: db.RabbitMQ,
			},
		},
		RedisRepo: repository.RedisRepository{
			Rdb: db.Redis,
		},
		LocalRepo: repository.LocalRepository{},
	}

	// 初始化订单处理器
	OrderHandler = handler.OrderHandler{
		OrderService: &service.OrderService{
			OrderRepo: repository.OrderRepository{
				DB: db.DB,
			},
		},
	}
}

func init() {
	initViper()
	db.InitDatabase()
	db.InitRedis()
	db.InitRabbitMQ()
	initHandler()
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 启动消息队列消费者
	go func() {
		receiver := receiver.NewReceiver(db.RabbitMQ, &repository.OrderRepository{DB: db.DB})
		if err := receiver.StartOrderConsumer(ctx); err != nil {
			log.Printf("启动订单消费者失败: %v", err)
		}
	}()

	// 初始化路由
	router := api.InitRouter(&UserHandler, &TicketHandler, &OrderHandler)

	// 获取端口配置
	port := viper.GetString("port")
	if port == "" {
		port = "8080"
	}

	// 启动HTTP服务器
	log.Printf("12305 票务系统启动成功！")
	log.Printf("服务器运行在端口: %s", port)
	log.Printf("访问地址: http://localhost:%s", port)
	log.Printf("API 文档:")
	log.Printf("   - 用户注册: POST http://localhost:%s/user/register", port)
	log.Printf("   - 用户登录: POST http://localhost:%s/user/login", port)
	log.Printf("   - 用户信息: GET http://localhost:%s/user/info", port)
	log.Printf("   - 票务列表: GET http://localhost:%s/ticket/list", port)
	log.Printf("   - 购买车票: POST http://localhost:%s/ticket/buy", port)
	log.Printf("   - 订单信息: GET http://localhost:%s/order/info", port)
	log.Printf("   - 订单支付: POST http://localhost:%s/order/pay", port)

	if err := router.Run(fmt.Sprintf(":%s", port)); err != nil && err != http.ErrServerClosed {
		log.Fatalf("启动服务器失败: %v", err)
	}
}
