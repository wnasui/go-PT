package handler

import (
	"12305/enum"
	"12305/model"
	"12305/repository"
	"12305/response"
	"12305/service"
	"12305/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type TicketHandler struct {
	ticketService service.TicketSrv
	redisRepo     repository.RedisRepository
	localRepo     repository.LocalRepository
}

func (h *TicketHandler) GetEntity(ticket model.Ticket) response.Ticket {
	return response.Ticket{
		ID:           utils.GetUUID(),
		Key:          utils.GetUUID(),
		TicketId:     ticket.TicketId,
		TicketNumber: ticket.TicketNumber,
		TicketTag:    ticket.TicketTag.String(),
		TicketPrice:  ticket.TicketPrice,
		CreateTime:   ticket.CreateTime,
		UpdateTime:   ticket.UpdateTime,
		DeleteTime:   ticket.DeleteTime,
	}
}

func (h *TicketHandler) TicketInfoHandler(c *gin.Context) {
	entity := response.Entity{
		Code:  int(enum.OperateOK),
		Msg:   enum.OperateOK.String(),
		Total: 0,
		Data:  nil,
	}

	ticketId := c.Param("ticket_id")
	if ticketId == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"entity": entity})
		return
	}
	ticket := model.Ticket{
		TicketId: ticketId,
	}
	result, err := h.ticketService.Get(c, &ticket)
	if err != nil {
		entity.Code = int(enum.OperateFailed)
		entity.Msg = enum.OperateFailed.String()
		c.JSON(http.StatusInternalServerError, gin.H{"entity": entity})
		return
	}

	r := h.GetEntity(*result)
	entity.Data = r
	entity.Msg = "success"
	c.JSON(http.StatusOK, gin.H{"entity": entity})
}

// 购买车票->本地扣减库存->本地生成订单->redis缓存扣减库存->MQ异步发送订单->数据库存储订单
func (h *TicketHandler) TicketBuyHandler(c *gin.Context) {
	entity := response.Entity{
		Code:  int(enum.OperateOK),
		Msg:   enum.OperateOK.String(),
		Total: 0,
		Data:  nil,
	}

	var ticket model.Ticket
	if err := c.ShouldBindJSON(&ticket); err != nil {
		entity.Code = int(enum.OperateFailed)
		entity.Msg = "参数错误: " + err.Error()
		c.JSON(http.StatusBadRequest, gin.H{"entity": entity})
		return
	}

	user, ok := c.Get("user")
	if !ok {
		entity.Code = int(enum.OperateFailed)
		entity.Msg = "用户信息获取失败"
		c.JSON(http.StatusUnauthorized, gin.H{"entity": entity})
		return
	}
	userInfo := user.(response.User)

	// 抢票
	result, err := h.ticketService.BuyTicketWriteThrough(c.Request.Context(), &ticket, userInfo)
	if err != nil {
		entity.Code = int(enum.OperateFailed)
		entity.Msg = "抢票失败: " + err.Error()
		c.JSON(http.StatusInternalServerError, gin.H{"entity": entity})
		return
	}

	if result {
		entity.Msg = "抢票成功"
		entity.Data = gin.H{
			"ticket_id": ticket.TicketId,
			"user_id":   userInfo.UserId,
			"status":    "success",
		}
		c.JSON(http.StatusOK, gin.H{"entity": entity})
	} else {
		entity.Code = int(enum.OperateFailed)
		entity.Msg = "抢票失败，票已售出或不可用"
		c.JSON(http.StatusConflict, gin.H{"entity": entity})
	}
}

// 缓存预热接口
func (h *TicketHandler) WarmUpCache(c *gin.Context) {
	ctx := c.Request.Context()

	err := h.ticketService.WarmUpCache(ctx)
	if err != nil {
		entity := response.Entity{
			Code: int(enum.OperateFailed),
			Msg:  "缓存预热失败: " + err.Error(),
		}
		c.JSON(http.StatusInternalServerError, gin.H{"entity": entity})
		return
	}

	entity := response.Entity{
		Code: int(enum.OperateOK),
		Msg:  "缓存预热完成",
	}
	c.JSON(http.StatusOK, gin.H{"entity": entity})
}

// 获取缓存统计接口
func (h *TicketHandler) GetCacheStats(c *gin.Context) {
	ctx := c.Request.Context()

	stats, err := h.ticketService.GetCacheStats(ctx)
	if err != nil {
		entity := response.Entity{
			Code: int(enum.OperateFailed),
			Msg:  "获取缓存统计失败: " + err.Error(),
		}
		c.JSON(http.StatusInternalServerError, gin.H{"entity": entity})
		return
	}

	entity := response.Entity{
		Code: int(enum.OperateOK),
		Msg:  "success",
		Data: stats,
	}
	c.JSON(http.StatusOK, gin.H{"entity": entity})
}

// Read-Through模式查询车票
func (h *TicketHandler) TicketListReadThroughHandler(c *gin.Context) {
	var trainnumber string
	entity := response.Entity{
		Code:      int(enum.OperateOK),
		Msg:       enum.OperateOK.String(),
		Total:     0,
		TotalPage: 1,
		Data:      nil,
	}

	if err := c.ShouldBindQuery(&trainnumber); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"entity": entity})
		return
	}

	tickets, err := h.ticketService.ListByTicketTagReadThrough(c, &trainnumber)
	if err != nil {
		entity.Code = int(enum.OperateFailed)
		entity.Msg = "查询失败: " + err.Error()
		c.JSON(http.StatusInternalServerError, gin.H{"entity": entity})
		return
	}

	entity.Data = tickets
	entity.Msg = "Read-Through模式查询成功"
	c.JSON(http.StatusOK, gin.H{"entity": entity})
}

// 查询抢票结果
func (h *TicketHandler) TicketBuyResultHandler(c *gin.Context) {
	taskId := c.Param("task_id")
	if taskId == "" {
		entity := response.Entity{
			Code: int(enum.OperateFailed),
			Msg:  "任务ID不能为空",
		}
		c.JSON(http.StatusBadRequest, gin.H{"entity": entity})
		return
	}

	// 这里应该从 Redis 或数据库中查询任务结果
	// 为了演示，我们返回一个模拟结果
	entity := response.Entity{
		Code: int(enum.OperateOK),
		Msg:  "查询成功",
		Data: gin.H{
			"task_id": taskId,
			"status":  "completed", // 可能的状态: processing, completed, failed
			"result":  "success",   // 可能的结果: success, failed
			"message": "抢票成功",
		},
	}

	c.JSON(http.StatusOK, gin.H{"entity": entity})
}

// // 用户搜索车票时获取所有同车次票
// func (h *TicketHandler) TicketListHandler(c *gin.Context) {
// 	//前端搜索车次，这里可以做一个哈希映射，让前端存储详细信息，后端根据结构体映射出tickettag
// 	//为了简单简化为直接传递tickettag
// 	var trainnumber string
// 	entity := response.Entity{
// 		Code:      int(enum.OperateOK),
// 		Msg:       enum.OperateOK.String(),
// 		Total:     0,
// 		TotalPage: 1,
// 		Data:      nil,
// 	}
// 	if err := c.ShouldBindQuery(&trainnumber); err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"entity": entity})
// 		return
// 	}

// 	Tickets, err := h.ticketService.ListByTicketTag(c, &trainnumber)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"entity": entity})
// 		return
// 	}
// 	entity.Data = Tickets
// 	entity.Msg = "success"
// 	c.JSON(http.StatusOK, gin.H{"entity": entity})
// }

// 异步抢票处理器
// func (h *TicketHandler) TicketBuyAsyncHandler(c *gin.Context) {
// 	entity := response.Entity{
// 		Code:  int(enum.OperateOK),
// 		Msg:   "抢票请求已提交，请稍后查询结果",
// 		Total: 0,
// 		Data:  nil,
// 	}

// 	var ticket model.Ticket
// 	if err := c.ShouldBindJSON(&ticket); err != nil {
// 		entity.Code = int(enum.OperateFailed)
// 		entity.Msg = "参数错误: " + err.Error()
// 		c.JSON(http.StatusBadRequest, gin.H{"entity": entity})
// 		return
// 	}

// 	user, ok := c.Get("user")
// 	if !ok {
// 		entity.Code = int(enum.OperateFailed)
// 		entity.Msg = "用户信息获取失败"
// 		c.JSON(http.StatusUnauthorized, gin.H{"entity": entity})
// 		return
// 	}
// 	userInfo := user.(response.User)

// 	// 生成任务ID
// 	taskId := utils.GetUUID()

// 	// 立即返回任务ID
// 	entity.Data = gin.H{
// 		"task_id":   taskId,
// 		"ticket_id": ticket.TicketId,
// 		"status":    "processing",
// 	}
// 	c.JSON(http.StatusAccepted, gin.H{"entity": entity})

// 	// 异步执行抢票任务
// 	go func() {
// 		// 创建新的 context，避免传递 gin.Context
// 		ctx := context.Background()

// 		result, err := h.ticketService.BuyTicketWriteThrough(ctx, &ticket, userInfo)

// 		// 这里可以将结果存储到 Redis 或数据库中
// 		// 供客户端通过 taskId 查询结果
// 		fmt.Printf("异步抢票任务完成 - TaskID: %s, 结果: %v, 错误: %v\n", taskId, result, err)

// 		// 可以发送 WebSocket 消息或推送通知给客户端
// 		// 这里只是打印日志，实际应用中应该实现结果通知机制
// 	}()
// }
