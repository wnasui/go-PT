package service

import (
	"12305/enum"
	"12305/model"
	"12305/repository"
	"12305/utils"
	"context"
	"fmt"
	"time"
)

type TicketService struct {
	ticketRepo repository.TicketRepository
	redisRepo  repository.RedisRepository
}

type TicketSrv interface {
	ListByTicketTag(ctx context.Context, trainnumber interface{}) ([]*model.Ticket, error)
	// GetTotal(ctx context.Context, req *query.ListQuery) (int64, error)
	Get(ctx context.Context, ticket *model.Ticket) (*model.Ticket, error)
	GetByTicketTag(ctx context.Context, TicketTag []enum.TicketTag) ([]*model.Ticket, error)
	GetByTicketNumber(ctx context.Context, seat int) (*model.Ticket, error)
	Exist(ctx context.Context, ticket *model.Ticket) (bool, error)
	Create(ctx context.Context, ticket *model.Ticket) (*model.Ticket, error)
	Edit(ctx context.Context, ticket *model.Ticket) (bool, error)
	Delete(ctx context.Context, ticket *model.Ticket) (bool, error)
}

func (s *TicketService) ListByTicketTag(ctx context.Context, trainnumber interface{}) ([]*model.Ticket, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	//通过本地缓存获取所有车票

	//通过redis获取所有同车次票
	tickettag := trainnumber.(string)
	return s.redisRepo.GetByTicketTag(ctx, tickettag)
}

// func (s *TicketService) GetTotal(req *query.ListQuery) (int64, error) {
// 	return s.ticketRepo.GetTotal(req)
// }

func (s *TicketService) Get(ctx context.Context, ticket *model.Ticket) (*model.Ticket, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return s.ticketRepo.Get(ctx, ticket)
}

func (s *TicketService) Exist(ctx context.Context, ticket *model.Ticket) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	exist, err := s.ticketRepo.Exist(ctx, *ticket)
	if err != nil {
		fmt.Println("查询车票是否存在失败", err)
		return false, err
	}
	if !exist {
		fmt.Println("车票不存在")
		return false, nil
	}
	return exist, nil
}

func (s *TicketService) Create(ctx context.Context, ticket *model.Ticket) (*model.Ticket, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	exist, err := s.ticketRepo.Exist(ctx, *ticket)
	if err != nil {
		fmt.Println("查询车票是否存在失败", err)
		return nil, err
	}
	if exist {
		fmt.Println("车票已存在")
		return nil, nil
	}
	Ticket, err := s.ticketRepo.Get(ctx, ticket)
	if err != nil {
		fmt.Println("获取车票失败", err)
		return nil, err
	}
	if Ticket.TicketId == "" {
		Ticket.TicketId = utils.GetUUID()
	}
	Ticket.CreateTime = time.Now()
	Ticket.UpdateTime = time.Now()
	Ticket.TicketTag = ticket.TicketTag
	Ticket.TicketNumber = ticket.TicketNumber
	Ticket.TicketPrice = ticket.TicketPrice
	if Ticket.TicketPrice == 0 {
		Ticket.TicketPrice = 999.999
	}
	return s.ticketRepo.CreateTicket(ctx, Ticket)
}

func (s *TicketService) Edit(ctx context.Context, ticket *model.Ticket) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	exist, err := s.ticketRepo.Exist(ctx, *ticket)
	if err != nil {
		fmt.Println("查询车票是否存在失败", err)
		return false, err
	}
	if !exist {
		fmt.Println("车票不存在")
		return false, nil
	}
	Ticket, err := s.ticketRepo.Get(ctx, ticket)
	if err != nil {
		fmt.Println("获取车票失败", err)
		return false, err
	}
	Ticket.TicketTag = ticket.TicketTag
	Ticket.TicketNumber = ticket.TicketNumber
	Ticket.TicketPrice = ticket.TicketPrice
	Ticket.UpdateTime = time.Now()
	return s.ticketRepo.Edit(ctx, Ticket)
}

func (s *TicketService) Delete(ctx context.Context, ticket *model.Ticket) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	exist, err := s.ticketRepo.Exist(ctx, *ticket)
	if err != nil {
		fmt.Println("查询车票是否存在失败", err)
		return false, err
	}
	if !exist {
		fmt.Println("车票不存在")
		return false, nil
	}
	return s.ticketRepo.Delete(ctx, ticket)
}
