package response

import (
	"12305/enum"
	"time"
)

type User struct {
	ID           string    `json:"id"`
	Key          string    `json:"key"`
	UserId       string    `json:"user_id"`
	UserIdentity string    `json:"user_identity"`
	UserPhone    string    `json:"user_phone"`
	UserName     string    `json:"user_name"`
	UserPwd      string    `json:"user_pwd"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Ticket struct {
	ID           string            `json:"id"`
	Key          string            `json:"key"`
	TicketId     string            `json:"ticket_id"`
	TicketNumber int               `json:"ticket_number"`
	TicketTag    string            `json:"ticket_tag"`
	TicketPrice  float64           `json:"ticket_price"`
	TicketStatus enum.TicketStatus `json:"ticket_status"`
	CreateTime   time.Time         `json:"create_time"`
	UpdateTime   time.Time         `json:"update_time"`
	DeleteTime   time.Time         `json:"delete_time"`
}

type Order struct {
	ID          string           `json:"id"`
	Key         string           `json:"key"`
	OrderId     string           `json:"order_id"`
	User        User             `json:"user"`
	Ticket      Ticket           `json:"ticket"`
	TotalPrice  int              `json:"total_price"`
	OrderStatus enum.OrderStatus `json:"order_status"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

type Entity struct {
	Code      int         `json:"code"`
	Msg       string      `json:"msg"`
	Total     int         `json:"total"`
	TotalPage int         `json:"total_page"`
	Data      interface{} `json:"data"`
}
