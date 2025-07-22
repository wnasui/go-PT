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
	// 本地缓存不存在扣减，只读缓存
	RefreshCache(ctx context.Context, tickettag string) error
	InvalidateCache(ctx context.Context, tickettag string) error
}

var localCache = utils.GetCache()

// 本地获取同车次票（只读）
func (repo *LocalRepository) GetByTicketTag(ctx context.Context, tickettag string) ([]*model.Ticket, error) {
	value, err := localCache.Get(ctx, tickettag)
	if err != nil {
		// 本地缓存未命中，从Redis加载
		return repo.refreshFromRedis(ctx, tickettag)
	}
	return value, nil
}

// refreshFromRedis 从Redis刷新本地缓存
func (repo *LocalRepository) refreshFromRedis(ctx context.Context, tickettag string) ([]*model.Ticket, error) {
	tickets, err := repo.RedisRepo.GetByTicketTag(ctx, tickettag)
	if err != nil {
		return nil, err
	}

	// 更新本地缓存
	if len(tickets) > 0 {
		localCache.Set(ctx, tickettag, tickets, 30*time.Second) // 本地缓存30秒
	}

	return tickets, nil
}

// RefreshCache 刷新缓存
func (repo *LocalRepository) RefreshCache(ctx context.Context, tickettag string) error {
	_, err := repo.refreshFromRedis(ctx, tickettag)
	return err
}

// InvalidateCache 使缓存失效
func (repo *LocalRepository) InvalidateCache(ctx context.Context, tickettag string) error {
	return localCache.Del(ctx, tickettag)
}
