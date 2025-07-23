package repository

import (
	"12305/model"
	"12305/utils"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

// 全局布隆过滤器实例
var bloomFilter *utils.BloomFilter

// 全局限流器实例
var rateLimiter *utils.RateLimiter

// 初始化防护组件
func init() {
	// 初始化布隆过滤器：大小10000，哈希函数数量5
	bloomFilter = utils.NewBloomFilter(10000, 5)

	// 初始化限流器：每秒100个请求，突发200个
	rateLimiter = utils.NewRateLimiter(time.Second, 200)
}

type RedisRepository struct {
	Rdb *redis.Client
	DB  *gorm.DB
}

type RedisRepoInterface interface {
	GetByTicketTag(ctx context.Context, tickettag string) ([]*model.Ticket, error)
	DecrStock(ctx context.Context, ticket *model.Ticket, remotstock *model.RemotStock) (*model.RemotStock, error)
	AddByTicketTag(ctx context.Context, tickettag string, remotstock *model.RemotStock) (*model.RemotStock, error)
	// 新增：分布式锁和库存一致性管理
	AcquireTicketLock(ctx context.Context, ticketId string, expireTime time.Duration) (bool, error)
	ReleaseTicketLock(ctx context.Context, ticketId string) error
	SyncTicketToCache(ctx context.Context, ticket *model.Ticket) error
	InvalidateTicketCache(ctx context.Context, ticketTag string) error
	// 新增：缓存统计
	GetCacheStats(ctx context.Context) (map[string]interface{}, error)
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

// 同步票务信息到缓存 - Write-Through模式（防雪崩）
func (repo *RedisRepository) SyncTicketToCache(ctx context.Context, ticket *model.Ticket) error {
	jsonData, err := json.Marshal(ticket)
	if err != nil {
		return fmt.Errorf("序列化票务数据失败: %v", err)
	}

	key := string(ticket.TicketTag)
	pipe := repo.Rdb.Pipeline()
	pipe.HSet(ctx, key, ticket.TicketId, jsonData)

	// 防雪崩：使用随机过期时间
	protector := utils.GetCacheProtector()
	randomExpiration := protector.GetRandomExpiration(30 * time.Minute)
	pipe.Expire(ctx, key, randomExpiration)

	//执行批量操作
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("更新Redis缓存失败: %v", err)
	}

	// 添加到布隆过滤器
	bloomFilter.Add(key)

	return nil
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

// 查询redis,如果redis中没有该票则查询mysql,并添加所有同车次票到redis（防击穿版）
func (repo *RedisRepository) DecrStock(ctx context.Context, ticket *model.Ticket, remotstock *model.RemotStock) (*model.RemotStock, error) {
	// 1. 限流检查
	if !rateLimiter.Allow() {
		return nil, fmt.Errorf("系统繁忙，请稍后重试")
	}

	// 2. 防穿透：检查布隆过滤器
	ticketKey := fmt.Sprintf("%s_%s", string(ticket.TicketTag), ticket.TicketId)
	if !bloomFilter.MayContain(ticketKey) {
		return nil, fmt.Errorf("票务信息不存在")
	}

	// 3. 防击穿：获取互斥锁
	protector := utils.GetCacheProtector()
	mutex := protector.GetMutex(ticketKey)
	mutex.Lock()
	defer mutex.Unlock()

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
			// 防穿透：缓存空值
			protector.CacheNull(ticketKey, 5*time.Minute)
			return nil, err
		}
		//添加到redis
		repo.AddByTicketTag(ctx, string(ticket.TicketTag), remotstock)
		// 添加到布隆过滤器
		bloomFilter.Add(ticketKey)
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

// 通过redis获取所有同车次票 (防击穿、防穿透、防雪崩)
func (repo *RedisRepository) GetByTicketTag(ctx context.Context, tickettag string) ([]*model.Ticket, error) {
	// 限流检查
	if !rateLimiter.Allow() {
		return nil, fmt.Errorf("系统繁忙，请稍后重试")
	}

	// 防穿透：检查布隆过滤器
	if !bloomFilter.MayContain(tickettag) {
		return nil, fmt.Errorf("车次不存在")
	}

	// 防穿透：检查空值缓存
	protector := utils.GetCacheProtector()
	if protector.IsNullCached(tickettag) {
		return nil, fmt.Errorf("车次暂无票务信息")
	}

	// 防击穿：获取互斥锁
	mutex := protector.GetMutex(tickettag)
	mutex.Lock()
	defer mutex.Unlock()

	// 标记热点数据
	protector.MarkHotKey(tickettag)

	TicketList := []*model.Ticket{}
	Ticketid := repo.Rdb.HGetAll(ctx, tickettag).Val()

	if len(Ticketid) == 0 {
		// 缓存空值，防止穿透
		protector.CacheNull(tickettag, 5*time.Minute)
		return nil, fmt.Errorf("车次暂无票务信息")
	}

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

// 获取Redis缓存统计信息
func (repo *RedisRepository) GetCacheStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 获取Redis信息
	info, err := repo.Rdb.Info(ctx, "memory").Result()
	if err != nil {
		return nil, err
	}

	// 获取数据库大小
	dbSize, err := repo.Rdb.DBSize(ctx).Result()
	if err != nil {
		return nil, err
	}

	// 获取连接数
	clientList, err := repo.Rdb.ClientList(ctx).Result()
	if err != nil {
		return nil, err
	}

	stats["info"] = info
	stats["db_size"] = dbSize
	stats["client_count"] = len(strings.Split(clientList, "\n"))

	return stats, nil
}
