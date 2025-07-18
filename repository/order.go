package repository

import (
	"12305/model"
	"12305/query"
	"12305/utils"
	"context"
	"time"

	"gorm.io/gorm"
)

type OrderRepository struct {
	DB *gorm.DB
}

type OrderRepoInterface interface {
	List(ctx context.Context, req *query.ListQuery) ([]*model.Order, error)
	GetTotal(ctx context.Context, req *query.ListQuery) (int64, error)
	Get(ctx context.Context, order model.Order) (*model.Order, error)
	Exist(ctx context.Context, order model.Order) (bool, error)
	CreateOrder(ctx context.Context, order *model.Order) (*model.Order, error)
	Edit(ctx context.Context, order *model.Order) (bool, error)
	Delete(ctx context.Context, order *model.Order) (bool, error)
}

func (repo *OrderRepository) List(ctx context.Context, req *query.ListQuery) ([]*model.Order, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	db := repo.DB
	limit, offset := utils.GetLimitAndOffset(req.Page, req.PageSize)
	var orders []*model.Order
	err := db.Order("id desc").Limit(limit).Offset(offset).Find(&orders).Error
	if err != nil {
		return nil, err
	}
	return orders, nil
}

func (repo *OrderRepository) GetTotal(ctx context.Context, req *query.ListQuery) (int64, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	db := repo.DB
	var total int64
	err := db.Find(&model.Order{}).Count(&total).Error
	if err != nil {
		return 0, err
	}
	return total, nil
}

func (repo *OrderRepository) Get(ctx context.Context, order model.Order) (*model.Order, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	db := repo.DB
	Order := model.Order{}
	err := db.Where("order_id=?", order.OrderId).First(&Order).Error
	if err != nil {
		return nil, err
	}
	return &Order, nil
}

func (repo *OrderRepository) Exist(ctx context.Context, order model.Order) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	db := repo.DB
	err := db.Where("order_id=?", order.OrderId).First(&order).Error
	if err != nil {
		return false, err
	}
	return true, nil
}

func (repo *OrderRepository) CreateOrder(ctx context.Context, order *model.Order) (*model.Order, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	db := repo.DB
	err := db.Create(order).Error
	if err != nil {
		return nil, err
	}
	return order, nil
}

func (repo *OrderRepository) Edit(ctx context.Context, order *model.Order) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	db := repo.DB
	err := db.Model(&order).Where("order_id=?", order.OrderId).Updates(map[string]interface{}{
		"order_status": order.OrderStatus,
		"update_time":  time.Now(),
		"total_price":  order.TotalPrice,
		"user":         order.User,
		"ticket":       order.Ticket,
	}).Error
	if err != nil {
		return false, err
	}
	return true, nil
}

func (repo *OrderRepository) Delete(ctx context.Context, order model.Order) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	db := repo.DB
	err := db.Model(&order).Where("order_id=?", order.OrderId).Delete(&order).Error
	if err != nil {
		return false, err
	}
	return true, nil
}
