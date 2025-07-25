package api

import (
	"12305/api/handler"

	"github.com/gin-gonic/gin"
)

func InitRouter(UserHandler *handler.UserHandler, TicketHandler *handler.TicketHandler, OrderHandler *handler.OrderHandler) *gin.Engine {
	router := gin.Default()
	//router.Use(cors.Default())//跨域
	router.Use(gin.Recovery())
	router.Use(gin.Logger())
	//router.Use(middleware.JwtAuth()) //事实上涉及支付使用session+cookie更安全

	// 用户相关路由
	userGroup := router.Group("/user")
	{
		userGroup.POST("/register", UserHandler.UserCreateHandler)
		userGroup.POST("/login", UserHandler.UserLoginHandler)
		userGroup.GET("/info", UserHandler.UserInfoHandler)
		userGroup.PUT("/edit", UserHandler.UserEditHandler)
		userGroup.DELETE("/delete", UserHandler.UserDeleteHandler)
	}

	// 票务相关路由
	ticketGroup := router.Group("/ticket")
	{
		ticketGroup.GET("/list", TicketHandler.TicketListReadThroughHandler)
		ticketGroup.POST("/buy", TicketHandler.TicketBuyHandler)
		// 新增：缓存管理路由
		ticketGroup.POST("/cache/warmup", TicketHandler.WarmUpCache)
		ticketGroup.GET("/cache/stats", TicketHandler.GetCacheStats)
	}

	// 订单相关路由
	orderGroup := router.Group("/order")
	{
		orderGroup.GET("/info", OrderHandler.OrderInfoHandler)
		orderGroup.POST("/pay", OrderHandler.OrderPayHandler)
	}

	return router
}
