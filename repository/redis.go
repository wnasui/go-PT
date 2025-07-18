package repository

import (
	"12305/model"
	"context"
	"encoding/json"
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
