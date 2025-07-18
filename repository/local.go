package repository

import (
	"12305/model"
	"12305/utils"
	"context"
	"time"

	"gorm.io/gorm"
)

type LocalRepository struct {
	DB        *gorm.DB
	RedisRepo *RedisRepository
}

type LocalRepoInterface interface {
	Get(ctx context.Context, key string) (interface{}, error)
	GetByTicketTag(ctx context.Context, tickettag string) ([]*model.Ticket, error)
	DecrStock(ctx context.Context, ticket *model.Ticket, remotstock *model.RemotStock) (*model.LocalStock, error)
	AddByTicketTag(ctx context.Context, tickettag string, remotstock *model.RemotStock) (*model.Ticket, error)
}

var localCache = utils.GetCache()

// 本地获取同车次票
func (repo *LocalRepository) GetByTicketTag(ctx context.Context, tickettag string) ([]*model.Ticket, error) {
	value, err := localCache.Get(ctx, tickettag)
	if err != nil {
		return nil, err
	}
	//若本地缓存没有则查找cache
	if value == nil {
		repo.RedisRepo.GetByTicketTag(ctx, tickettag)
	}
	return value, nil
}

// 本地扣减库存
func (repo *LocalRepository) DecrStock(ctx context.Context, ticket *model.Ticket, localstock *model.LocalStock) (*model.LocalStock, error) {
	TicketSlice, err := localCache.Get(ctx, string(ticket.TicketTag))
	if err != nil {
		return nil, err
	}
	//若本地缓存没有则查找redis
	if len(TicketSlice) == 0 {
		repo.RedisRepo.GetByTicketTag(ctx, string(ticket.TicketTag))
		return nil, nil
	} else {
		for i, Ticket := range TicketSlice {
			if Ticket.TicketId == ticket.TicketId {
				TicketSlice = append(TicketSlice[:i], TicketSlice[i+1:]...)
			}
		}
		localstock.LocalStockNum = len(TicketSlice)
		localCache.Set(ctx, string(ticket.TicketTag), TicketSlice, 10*time.Second)
		return localstock, nil
	}
}
