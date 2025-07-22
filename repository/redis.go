package repository

import (
	"12305/model"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type RedisRepository struct {
	Rdb *redis.Client
	DB  *gorm.DB
}

type RedisRepoInterface interface {
	GetByTicketTag(ctx context.Context, tickettag string) ([]*model.Ticket, error)
	DecrStock(ctx context.Context, ticket *model.Ticket, remotstock *model.RemotStock) (*model.RemotStock, error)
	AddByTicketTag(ctx context.Context, tickettag string, remotstock *model.RemotStock) (*model.Ticket, error)
	// 新增：分布式锁和库存一致性管理
	AcquireTicketLock(ctx context.Context, ticketId string, expireTime time.Duration) (bool, error)
	ReleaseTicketLock(ctx context.Context, ticketId string) error
	SyncTicketToCache(ctx context.Context, ticket *model.Ticket) error
	InvalidateTicketCache(ctx context.Context, ticketTag string) error
}

// 获取票务分布式锁
func (repo *RedisRepository) AcquireTicketLock(ctx context.Context, ticketId string, expireTime time.Duration) (bool, error) {
	lockKey := fmt.Sprintf("ticket_lock_%s", ticketId)

	// 使用SET NX EX命令实现分布式锁
	result, err := repo.Rdb.SetNX(ctx, lockKey, "1", expireTime).Result()
	if err != nil {
		return false, err
	}

	return result, nil
}

// 释放票务分布式锁
func (repo *RedisRepository) ReleaseTicketLock(ctx context.Context, ticketId string) error {
	lockKey := fmt.Sprintf("ticket_lock_%s", ticketId)

	// 使用Lua脚本确保原子性释放锁
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`

	_, err := repo.Rdb.Eval(ctx, script, []string{lockKey}, []string{"1"}).Result()
	return err
}

// 同步票务信息到缓存
func (repo *RedisRepository) SyncTicketToCache(ctx context.Context, ticket *model.Ticket) error {
	jsonData, err := json.Marshal(ticket)
	if err != nil {
		return err
	}

	// 使用Hash结构存储，便于批量操作
	key := string(ticket.TicketTag)
	return repo.Rdb.HSet(ctx, key, ticket.TicketId, jsonData).Err()
}

// 使票务缓存失效
func (repo *RedisRepository) InvalidateTicketCache(ctx context.Context, ticketTag string) error {
	// 删除整个车次的缓存，强制重新加载
	return repo.Rdb.Del(ctx, ticketTag).Err()
}

// 采用根据车次缓存票（查redis如果没有该票则拉取所有同车次的票到redis）

// 拉取所有同车次的票到redis
func (repo *RedisRepository) AddByTicketTag(ctx context.Context, tickettag string, remotstock *model.RemotStock) (*model.RemotStock, error) {
	db := repo.DB
	var Ticket []model.Ticket
	err := db.Where("ticket_tag=?", tickettag).Find(&Ticket).Error
	if err != nil {
		return nil, err
	}
	for _, ticket := range Ticket {
		jsonData, err := json.Marshal(ticket)
		if err != nil {
			return nil, err
		}
		repo.Rdb.HSet(ctx, tickettag, ticket.TicketId, jsonData, 10*time.Second)
	}
	remotstocknum, err := repo.Rdb.HLen(ctx, tickettag).Result()
	if err != nil {
		return nil, err
	}
	remotstock.RemotStockNum = int(remotstocknum)
	return remotstock, nil
}

// 查询redis,如果redis中没有该票则查询mysql,并添加所有同车次票到redis
func (repo *RedisRepository) DecrStock(ctx context.Context, ticket *model.Ticket, remotstock *model.RemotStock) (*model.RemotStock, error) {
	var db = repo.DB
	Ticket, err := repo.Rdb.HGet(ctx, string(ticket.TicketTag), ticket.TicketId).Result()
	if err != nil {
		return nil, err
	}
	if Ticket == "" {
		//未找到则查询mysql
		TicketModel := model.Ticket{}
		err = db.Where("ticket_id=?", ticket.TicketId).First(&TicketModel).Error
		if err != nil {
			return nil, err
		}
		//添加到redis
		repo.AddByTicketTag(ctx, string(ticket.TicketTag), remotstock)
	} else { //查询到则削减库存
		repo.Rdb.HDel(ctx, string(ticket.TicketTag), ticket.TicketId)
		remotstocknum, err := repo.Rdb.HLen(ctx, string(ticket.TicketTag)).Result()
		if err != nil {
			return nil, err
		}
		remotstock.RemotStockNum = int(remotstocknum)
	}
	return remotstock, nil
}

// 通过redis获取所有同车次票
func (repo *RedisRepository) GetByTicketTag(ctx context.Context, tickettag string) ([]*model.Ticket, error) {
	TicketList := []*model.Ticket{}
	Ticketid := repo.Rdb.HGetAll(ctx, tickettag).Val()
	for _, value := range Ticketid {
		var ticket model.Ticket
		err := json.Unmarshal([]byte(value), &ticket)
		if err != nil {
			return nil, err
		}
		TicketList = append(TicketList, &ticket)
	}
	return TicketList, nil
}
