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

// 用户搜索车票时获取所有同车次票
func (h *TicketHandler) TicketListHandler(c *gin.Context) {
	//前端搜索车次，这里可以做一个哈希映射，让前端存储详细信息，后端根据结构体映射出tickettag
	//为了简单简化为直接传递tickettag
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

	Tickets, err := h.ticketService.ListByTicketTag(c, &trainnumber)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"entity": entity})
		return
	}
	entity.Data = Tickets
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
		c.JSON(http.StatusInternalServerError, gin.H{"entity": entity})
		return
	}
	user, ok := c.Get("user")
	if !ok {
		entity.Code = int(enum.OperateFailed)
		entity.Msg = enum.OperateFailed.String()
		c.JSON(http.StatusInternalServerError, gin.H{"entity": entity})
		return
	}
	userInfo := user.(response.User)
	//异步抢票
	go func() {
		result, err := h.ticketService.BuyTicket(c, &ticket, userInfo)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"entity": entity})
			return
		}
		if result {
			entity.Msg = "success"
		} else {
			entity.Msg = "failed"
		}
		c.JSON(http.StatusOK, gin.H{"entity": entity})
	}()
	//跳转支付页面
	c.Redirect(http.StatusFound, "/pay")
}
