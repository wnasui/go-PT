package model

import "12305/enum"

// 本地库存
type LocalStock struct {
	TicketTag         enum.TicketTag `json:"ticket_tag"`
	LocalStockNum     int            `json:"local_stock_num"`
	LocalStockSoldNum int            `json:"local_stock_sold_num"`
}

// 远程库存
type RemotStock struct {
	TicketTag         enum.TicketTag `json:"ticket_tag" gorm:"column:ticket_tag"`
	RemotStockNum     int            `json:"remot_stock_num" gorm:"column:remot_stock_num"`
	RemotStockSoldNum int            `json:"remot_stock_sold_num" gorm:"column:remot_stock_sold_num"`
}

//buffer库存,作为容错
type BufferStock struct {
	TicketTag          enum.TicketTag `json:"ticket_tag" gorm:"column:ticket_tag"`
	BufferStockNum     int            `json:"buffer_stock_num" gorm:"column:buffer_stock_num"`
	BufferStockSoldNum int            `json:"buffer_stock_sold_num" gorm:"column:buffer_stock_sold_num"`
}
