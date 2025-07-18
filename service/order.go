package service

import (
	"12305/enum"
	"12305/model"
	"12305/query"
	"12305/repository"
	"12305/utils"
	"context"
	"fmt"
	"time"
)

type OrderService struct {
	orderRepo repository.OrderRepository
}

type OrderSrv interface {
	List(ctx context.Context, req *query.ListQuery) ([]*model.Order, error)
	GetTotal(ctx context.Context, req *query.ListQuery) (int64, error)
	Get(ctx context.Context, order *model.Order) (*model.Order, error)
	Exist(ctx context.Context, order *model.Order) (bool, error)
	Create(ctx context.Context, order *model.Order) (*model.Order, error)
	Edit(ctx context.Context, order *model.Order) (bool, error)
	Delete(ctx context.Context, order *model.Order) (bool, error)
}

func (s *OrderService) List(ctx context.Context, req *query.ListQuery) ([]*model.Order, error) {
	return s.orderRepo.List(ctx, req)
}

func (s *OrderService) GetTotal(ctx context.Context, req *query.ListQuery) (int64, error) {
	return s.orderRepo.GetTotal(ctx, req)
}

func (s *OrderService) Get(ctx context.Context, order *model.Order) (*model.Order, error) {
	exist, err := s.orderRepo.Exist(ctx, *order)
	if err != nil {
		return nil, err
	}
	if !exist {
		fmt.Println("订单不存在")
		return nil, nil
	}
	return s.orderRepo.Get(ctx, *order)
}

func (s *OrderService) Exist(ctx context.Context, order *model.Order) (bool, error) {
	return s.orderRepo.Exist(ctx, *order)
}

func (s *OrderService) Create(ctx context.Context, order *model.Order) (*model.Order, error) {
	exist, err := s.orderRepo.Exist(ctx, *order)
	if err != nil {
		return nil, err
	}
	if exist {
		fmt.Println("订单已存在")
		return nil, nil
	}
	if order.OrderId == "" {
		order.OrderId = utils.GetUUID()
	}
	order.CreateTime = time.Now()
	order.UpdateTime = time.Now()
	order.OrderStatus = enum.OrderStatusNormal
	return s.orderRepo.CreateOrder(ctx, order)
}
