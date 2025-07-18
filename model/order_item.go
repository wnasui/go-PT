package model

import "time"

type OrderItem struct {
	OrderItemId   string    `json:"order_item_id"`
	TicketId      string    `json:"ticket_id"`
	Quantity      int       `json:"quantity"`
	TotalPrice    float64   `json:"total_price"`
	Departure     string    `json:"departure"`
	Destination   string    `json:"destination"`
	DepartureTime time.Time `json:"departure_time"`
	ArrivalTime   time.Time `json:"arrival_time"`
	CreateTime    time.Time `json:"create_at"`
	UpdateTime    time.Time `json:"update_at"`
}
