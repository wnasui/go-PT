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
	OrderRepo repository.OrderRepository
}

type OrderSrv interface {
	List(ctx context.Context, req *query.ListQuery) ([]*model.Order, error)
	GetTotal(ctx context.Context, req *query.ListQuery) (int64, error)
	Get(ctx context.Context, order *model.Order) (*model.Order, error)
	Exist(ctx context.Context, order *model.Order) (bool, error)
	Create(ctx context.Context, order *model.Order) (*model.Order, error)
	Edit(ctx context.Context, order *model.Order) (bool, error)
	Delete(ctx context.Context, order *model.Order) (bool, error)
	// 处理消息队列相关业务逻辑
	ProcessOrderFromMQ(ctx context.Context, order *model.Order) error
	ValidateOrderFromMQ(ctx context.Context, order *model.Order) error
}

func (s *OrderService) List(ctx context.Context, req *query.ListQuery) ([]*model.Order, error) {
	return s.OrderRepo.List(ctx, req)
}

func (s *OrderService) GetTotal(ctx context.Context, req *query.ListQuery) (int64, error) {
	return s.OrderRepo.GetTotal(ctx, req)
}

func (s *OrderService) Get(ctx context.Context, order *model.Order) (*model.Order, error) {
	exist, err := s.OrderRepo.Exist(ctx, order)
	if err != nil {
		return nil, err
	}
	if !exist {
		fmt.Println("订单不存在")
		return nil, nil
	}
	return s.OrderRepo.Get(ctx, *order)
}

func (s *OrderService) Exist(ctx context.Context, order *model.Order) (bool, error) {
	return s.OrderRepo.Exist(ctx, order)
}

func (s *OrderService) Create(ctx context.Context, order *model.Order) (*model.Order, error) {
	exist, err := s.OrderRepo.Exist(ctx, order)
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
	return s.OrderRepo.CreateOrder(ctx, order)
}

func (s *OrderService) Edit(ctx context.Context, order *model.Order) (bool, error) {
	exist, err := s.OrderRepo.Exist(ctx, order)
	if err != nil {
		return false, err
	}
	if !exist {
		fmt.Println("订单不存在")
		return false, nil
	}
	return s.OrderRepo.Edit(ctx, order)
}

func (s *OrderService) Delete(ctx context.Context, order *model.Order) (bool, error) {
	exist, err := s.OrderRepo.Exist(ctx, order)
	if err != nil {
		return false, err
	}
	if !exist {
		fmt.Println("订单不存在")
		return false, nil
	}
	return s.OrderRepo.Delete(ctx, order)
}

// ProcessOrderFromMQ 处理来自消息队列的订单（包含业务逻辑验证）
func (s *OrderService) ProcessOrderFromMQ(ctx context.Context, order *model.Order) error {
	// 1. 业务逻辑验证
	if err := s.ValidateOrderFromMQ(ctx, order); err != nil {
		return err
	}

	// 2. 设置默认值
	if order.CreateTime.IsZero() {
		order.CreateTime = time.Now()
	}
	order.UpdateTime = time.Now()

	// 3. 调用Repository层处理数据存储
	return s.OrderRepo.ProcessOrderFromMQ(ctx, order)
}

// ValidateOrderFromMQ 验证来自消息队列的订单数据
func (s *OrderService) ValidateOrderFromMQ(ctx context.Context, order *model.Order) error {
	// 验证订单ID
	if order.OrderId == "" {
		return fmt.Errorf("订单ID不能为空")
	}

	// 验证订单状态
	if order.OrderStatus == 0 {
		order.OrderStatus = enum.OrderStatusNormal
	}

	// 验证价格
	if order.TotalPrice <= 0 {
		return fmt.Errorf("订单总价必须大于0")
	}

	// 验证用户信息
	if order.User.UserId == "" {
		return fmt.Errorf("用户ID不能为空")
	}

	// 验证票务信息
	if order.Ticket.TicketId == "" {
		return fmt.Errorf("票务ID不能为空")
	}

	return nil
}
