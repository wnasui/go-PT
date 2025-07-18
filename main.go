package main

import (
	"12305/enum"
	"12305/model"
)

// 假设单机初始化后的本地库存、远程库存、buffer库存
const (
	LOCALSTOCK  = 1000
	REMOTSTOCK  = 10000
	BUFFERSTOCK = 500
	TICKETTAG   = enum.TicketTag(1)
)

var localstock = model.LocalStock{
	LocalStockNum:     LOCALSTOCK,
	LocalStockSoldNum: 0,
	TicketTag:         TICKETTAG,
}
var remotstock = model.RemotStock{
	RemotStockNum:     REMOTSTOCK,
	RemotStockSoldNum: 0,
	TicketTag:         TICKETTAG,
}
var bufferstock = model.BufferStock{
	BufferStockNum:     BUFFERSTOCK,
	BufferStockSoldNum: 0,
	TicketTag:          TICKETTAG,
}
