package api

import (
	"12305/api/handler"
	"12305/middleware"

	"github.com/gin-gonic/gin"
)

func InitRouter(UserHandler *handler.UserHandler, TicketHandler *handler.TicketHandler, OrderHandler *handler.OrderHandler) *gin.Engine {
	router := gin.Default()
	//router.Use(cors.Default())//跨域
	router.Use(gin.Recovery())
	router.Use(gin.Logger())
	router.Use(middleware.JwtAuth()) //事实上涉及支付使用session+cookie更安全

	router.Group("/user")
	{
		router.POST("/register", UserHandler.UserCreateHandler)
		router.POST("/login", UserHandler.UserLoginHandler)
		router.GET("/info", UserHandler.UserInfoHandler)
		router.PUT("/edit", UserHandler.UserEditHandler)
		router.DELETE("/delete", UserHandler.UserDeleteHandler)
	}

	router.Group("/ticket")
	{
		router.GET("/list", TicketHandler.TicketListReadThroughHandler)
		router.POST("/buy", TicketHandler.TicketBuyHandler)
		// 新增：缓存管理路由
		router.POST("/cache/warmup", TicketHandler.WarmUpCache)
		router.GET("/cache/stats", TicketHandler.GetCacheStats)
	}

	router.Group("/order")
	{
		router.GET("/info", OrderHandler.OrderInfoHandler)
		router.POST("/pay", OrderHandler.OrderPayHandler)
	}

	return router
}
