package handler

import (
	"12305/enum"
	"12305/model"
	"12305/response"
	"12305/service"
	"12305/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	OrderService service.OrderSrv
}

func (h *OrderHandler) GetEntity(order model.Order) response.Order {
	return response.Order{
		ID:          utils.GetUUID(),
		Key:         utils.GetUUID(),
		OrderId:     order.OrderId,
		User:        order.User,
		Ticket:      h.convertTicket(order.Ticket),
		TotalPrice:  order.TotalPrice,
		OrderStatus: order.OrderStatus,
		CreatedAt:   order.CreateTime,
		UpdatedAt:   order.UpdateTime,
	}
}

func (h *OrderHandler) convertTicket(ticket model.Ticket) response.Ticket {
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

func (h *OrderHandler) OrderInfoHandler(c *gin.Context) {
	entity := response.Entity{
		Code:  int(enum.OperateOK),
		Msg:   enum.OperateOK.String(),
		Total: 0,
		Data:  nil,
	}

	orderId := c.Param("order_id")
	if orderId == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"entity": entity})
		return
	}
	order := model.Order{
		OrderId: orderId,
	}

	result, err := h.OrderService.Get(c, &order)
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

// 支付
func (h *OrderHandler) OrderPayHandler(c *gin.Context) {
	entity := response.Entity{
		Code:  int(enum.OperateOK),
		Msg:   enum.OperateOK.String(),
		Total: 0,
		Data:  nil,
	}
	//支付相关
	c.JSON(http.StatusOK, gin.H{"entity": entity})
}
