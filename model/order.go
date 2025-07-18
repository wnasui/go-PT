package model

import (
	"12305/enum"
	"time"
)

type Order struct {
	OrderId     string           `json:"order_id" gorm:"column:order_id"`
	OrderStatus enum.OrderStatus `json:"order_status" gorm:"column:order_status"` //0:未支付，1：已支付，2：已退,3:已删除
	TotalPrice  float64          `json:"total_price" gorm:"column:total_price"`
	CreateTime  time.Time        `json:"create_at" gorm:"column:create_at"`
	UpdateTime  time.Time        `json:"update_at" gorm:"column:update_at"`
	DeleteTime  time.Time        `json:"delete_at" gorm:"column:delete_at"`
	User        User             `json:"user" gorm:"foreignKey:UserId"`
	Ticket      Ticket           `json:"ticket" gorm:"foreignKey:TicketId"`
}
