package repository

import (
	"12305/enum"
	"12305/model"
	"12305/query"
	"12305/utils"
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type TicketRepository struct {
	DB  *gorm.DB
	Rdb *redis.Client
}

type TicketRepoInterface interface {
	List(ctx context.Context, req *query.ListQuery) ([]*model.Ticket, error)
	// GetTotal(req *query.ListQuery) (int64, error)
	Get(ctx context.Context, ticket *model.Ticket) (*model.Ticket, error)
	GetByTicketTag(ctx context.Context, TicketTag enum.TicketTag) ([]*model.Ticket, error)
	GetByTicketNumber(ctx context.Context, seat int) (*model.Ticket, error)
	Exist(ctx context.Context, ticket model.Ticket) (bool, error)
	CreateTicket(ctx context.Context, ticket *model.Ticket) (*model.Ticket, error)
	Edit(ctx context.Context, ticket *model.Ticket) (bool, error)
	Delete(ctx context.Context, ticket *model.Ticket) (bool, error)
}

func (repo *TicketRepository) List(ctx context.Context, req *query.ListQuery) ([]*model.Ticket, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	db := repo.DB
	limit, offset := utils.GetLimitAndOffset(req.Page, req.PageSize)
	var tickets []*model.Ticket
	err := db.Order("ticket_tag desc").Limit(limit).Offset(offset).Find(&tickets).Error
	if err != nil {
		return nil, err
	}
	return tickets, nil
}

// func (repo *TicketRepository) GetTotal(req *query.ListQuery) (int64, error) {
// 	db := repo.DB
// 	var total int64
// 	err := db.Find(&model.Ticket{}).Count(&total).Error
// 	if err != nil {
// 		return 0, err
// 	}
// 	return total, nil
// }

func (repo *TicketRepository) Get(ctx context.Context, ticket *model.Ticket) (*model.Ticket, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	db := repo.DB
	Ticket := model.Ticket{}
	err := db.Where("ticket_id=?", ticket.TicketId).First(&Ticket).Error
	if err != nil {
		return nil, err
	}
	return &Ticket, nil
}

func (repo *TicketRepository) GetByTicketTag(ctx context.Context, TicketTag []enum.TicketTag) ([]*model.Ticket, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	db := repo.DB
	var tickets []*model.Ticket
	for i := range TicketTag {
		db.Where("TicketTag=?", TicketTag[i]).Find(&tickets)
	}
	return tickets, nil
}

func (repo *TicketRepository) GetByTicketNumber(ctx context.Context, seat int) (*model.Ticket, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	db := repo.DB
	Ticket := model.Ticket{}
	err := db.Where("ticket_number=?", seat).First(&Ticket).Error
	if err != nil {
		return nil, err
	}
	return &Ticket, nil
}

func (repo *TicketRepository) Exist(ctx context.Context, ticket model.Ticket) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	db := repo.DB
	err := db.Where("ticket_id=?", ticket.TicketId).First(&ticket).Error
	if ticket.TicketId == "" {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (repo *TicketRepository) CreateTicket(ctx context.Context, ticket *model.Ticket) (*model.Ticket, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	db := repo.DB
	err := db.Create(ticket).Error
	if err != nil {
		return nil, err
	}
	return ticket, nil
}

func (repo *TicketRepository) Edit(ctx context.Context, ticket *model.Ticket) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	db := repo.DB
	err := db.Model(&ticket).Where("ticket_id=?", ticket.TicketId).Updates(map[string]interface{}{
		"ticket_Tag":   ticket.TicketTag,
		"ticket_price": ticket.TicketPrice,
		"status":       ticket.TicketStatus,
		"update_time":  time.Now(),
	}).Error
	if err != nil {
		return false, err
	}
	return true, nil
}

func (repo *TicketRepository) Delete(ctx context.Context, ticket *model.Ticket) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	db := repo.DB
	err := db.Model(&ticket).Where("ticket_id=?", ticket.TicketId).Delete(&ticket).Error
	if err != nil {
		return false, err
	}
	return true, nil
}
