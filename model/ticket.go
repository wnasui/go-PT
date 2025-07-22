package model

import (
	"12305/enum"
	"time"
)

type Ticket struct {
	TicketId     string            `json:"ticket_id" gorm:"column:ticket_id;primaryKey"`
	TicketNumber int               `json:"ticket_number" gorm:"column:ticket_number"` //座位号，按照顺序编号
	TicketTag    enum.TicketTag    `json:"ticket_tag" gorm:"column:ticket_tag"`       //车次tag
	TicketPrice  float64           `json:"ticket_price" gorm:"column:ticket_price"`
	TicketStatus enum.TicketStatus `json:"status" gorm:"column:status"`             //0:未售，1：已售，2：已退, 3:已删除
	Version      int64             `json:"version" gorm:"column:version;default:0"` // 乐观锁版本号
	CreateTime   time.Time         `json:"create_at" gorm:"column:create_at"`
	UpdateTime   time.Time         `json:"update_at" gorm:"column:update_at"`
	DeleteTime   time.Time         `json:"delete_at" gorm:"column:delete_at"`
}
