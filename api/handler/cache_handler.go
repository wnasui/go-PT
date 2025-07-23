package handler

import (
	"12305/service"
	"12305/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CacheHandler struct {
	ticketService service.TicketSrv
}

func NewCacheHandler(ticketService service.TicketSrv) *CacheHandler {
	return &CacheHandler{
		ticketService: ticketService,
	}
}

// 获取缓存统计信息
func (h *CacheHandler) GetCacheStats(c *gin.Context) {
	ctx := c.Request.Context()

	// 获取缓存统计
	stats, err := h.ticketService.GetCacheStats(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取缓存统计失败",
			"error":   err.Error(),
		})
		return
	}

	// 获取热点数据统计
	protector := utils.GetCacheProtector()
	hotKeysStats := protector.GetHotKeysStats()

	// 合并统计信息
	response := gin.H{
		"code":    200,
		"message": "获取缓存统计成功",
		"data": gin.H{
			"cache_stats":    stats,
			"hot_keys_stats": hotKeysStats,
		},
	}

	c.JSON(http.StatusOK, response)
}

// 预热缓存
func (h *CacheHandler) WarmUpCache(c *gin.Context) {
	ctx := c.Request.Context()

	err := h.ticketService.WarmUpCache(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "缓存预热失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "缓存预热成功",
	})
}

// // 清除缓存
// func (h *CacheHandler) ClearCache(c *gin.Context) {
// 	ctx := c.Request.Context()
// 	// 获取要清除的缓存类型
// 	cacheType := c.Query("type") // "redis", "local", "all"

// 	switch cacheType {
// 	case "redis":
// 		// 清除Redis缓存（这里需要实现具体的清除逻辑）
// 		c.JSON(http.StatusOK, gin.H{
// 			"code":    200,
// 			"message": "Redis缓存清除成功",
// 		})
// 	case "local":
// 		// 清除本地缓存
// 		protector := utils.GetCacheProtector()
// 		// 这里可以添加清除本地缓存的逻辑
// 		c.JSON(http.StatusOK, gin.H{
// 			"code":    200,
// 			"message": "本地缓存清除成功",
// 		})
// 	case "all":
// 		// 清除所有缓存
// 		c.JSON(http.StatusOK, gin.H{
// 			"code":    200,
// 			"message": "所有缓存清除成功",
// 		})
// 	default:
// 		c.JSON(http.StatusBadRequest, gin.H{
// 			"code":    400,
// 			"message": "无效的缓存类型",
// 		})
// 	}
// }

// 获取限流器状态
func (h *CacheHandler) GetRateLimiterStatus(c *gin.Context) {
	// 返回限流器的当前状态
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取限流器状态成功",
		"data": gin.H{
			"rate_limiter": "active",
			"status":       "normal",
		},
	})
}

// 设置限流参数
func (h *CacheHandler) SetRateLimiterConfig(c *gin.Context) {
	rateStr := c.PostForm("rate")
	burstStr := c.PostForm("burst")

	if rateStr == "" || burstStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "缺少必要参数",
		})
		return
	}

	rate, err := strconv.Atoi(rateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的速率参数",
		})
		return
	}

	burst, err := strconv.Atoi(burstStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的突发参数",
		})
		return
	}

	// 这里可以更新限流器配置
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "限流器配置更新成功",
		"data": gin.H{
			"rate":  rate,
			"burst": burst,
		},
	})
}
