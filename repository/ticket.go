package repository

import (
	"12305/enum"
	"12305/model"
	"12305/query"
	"12305/utils"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
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
	//开启事务
	ExecuteTransaction(fn func(r *TicketRepository) error) error
	//乐观锁扣减库存
	DecreaseStockWithOptimisticLock(ctx context.Context, ticketId string, version int64) (bool, error)
	// 乐观锁更新票务状态
	UpdateTicketStatusWithOptimisticLock(ctx context.Context, ticketId string, oldVersion int64, newStatus enum.TicketStatus) (bool, error)
	// 带重试的乐观锁更新票务状态
	UpdateTicketStatusWithOptimisticLockRetry(ctx context.Context, ticketId string, newStatus enum.TicketStatus, maxRetries int) (bool, error)
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

func (repo *TicketRepository) ExecuteTransaction(fn func(r *TicketRepository) error) error {
	return repo.DB.Transaction(func(tx *gorm.DB) error {
		txTicketRepo := &TicketRepository{DB: tx}
		return fn(txTicketRepo)
	})
}

func (repo *TicketRepository) DecreaseStockWithOptimisticLock(ctx context.Context, ticketId string, version int64) (bool, error) {
	result := repo.DB.Model(&model.Ticket{}).
		Where("ticket_id = ? AND version = ? AND status = ?", ticketId, version, enum.TicketStatusNormal).
		Updates(map[string]interface{}{
			"status":      enum.TicketStatusSold,
			"version":     version + 1,
			"update_time": time.Now(),
		})

	if result.Error != nil {
		return false, result.Error
	}

	// 检查是否真的更新了记录
	return result.RowsAffected > 0, nil
}

// UpdateTicketStatusWithOptimisticLock 使用乐观锁更新票务状态
func (repo *TicketRepository) UpdateTicketStatusWithOptimisticLock(ctx context.Context, ticketId string, oldVersion int64, newStatus enum.TicketStatus) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}

	// 验证版本号合理性
	if oldVersion < 0 {
		return false, fmt.Errorf("无效的版本号: %d", oldVersion)
	}

	result := repo.DB.Model(&model.Ticket{}).
		Where("ticket_id = ? AND version = ? AND status = ?", ticketId, oldVersion, enum.TicketStatusNormal).
		Updates(map[string]interface{}{
			"status":      newStatus,
			"version":     oldVersion + 1,
			"update_time": time.Now(),
		})

	if result.Error != nil {
		return false, result.Error
	}

	// 检查是否真的更新了记录
	return result.RowsAffected > 0, nil
}

// UpdateTicketStatusWithOptimisticLockRetry 带重试的乐观锁更新
func (repo *TicketRepository) UpdateTicketStatusWithOptimisticLockRetry(ctx context.Context, ticketId string, newStatus enum.TicketStatus, maxRetries int) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}

	if maxRetries <= 0 {
		maxRetries = 3 // 默认重试3次
	}

	for i := 0; i < maxRetries; i++ {
		// 重新获取最新数据
		currentTicket, err := repo.Get(ctx, &model.Ticket{TicketId: ticketId})
		if err != nil {
			// 如果是最后一次重试，返回错误
			if i == maxRetries-1 {
				return false, fmt.Errorf("获取票务信息失败: %v", err)
			}
			// 否则继续重试
			continue
		}

		// 检查票务状态
		if currentTicket.TicketStatus != enum.TicketStatusNormal {
			return false, errors.New("票已售出或不可用")
		}

		// 检查版本号异常
		if currentTicket.Version < 0 {
			return false, fmt.Errorf("票务版本号异常: %d", currentTicket.Version)
		}

		// 尝试乐观锁更新
		success, err := repo.UpdateTicketStatusWithOptimisticLock(ctx, ticketId, currentTicket.Version, newStatus)
		if err != nil {
			// 如果是最后一次重试，返回错误
			if i == maxRetries-1 {
				return false, err
			}
			// 否则继续重试
			continue
		}

		if success {
			return true, nil
		}

		// 如果失败且还有重试次数，等待一小段时间后重试
		if i < maxRetries-1 {
			select {
			case <-time.After(10 * time.Millisecond):
			case <-ctx.Done():
				return false, ctx.Err()
			}
		}
	}

	return false, nil
}
