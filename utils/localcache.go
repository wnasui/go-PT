package utils

import (
	"12305/model"
	"context"
	"errors"
	"sync"
	"time"
)

// 本地cache       //map在并发情况下不安全，需要使用sync.Map
type Cache struct {
	data sync.Map
	mu   sync.RWMutex
}

type LocalCache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, expiration time.Duration) error
	Del(ctx context.Context, key string) error
	Len(ctx context.Context, key string) (int, error)
	Edit(ctx context.Context, key string, value []*model.Ticket) error
	// 新增：获取缓存统计
	GetStats(ctx context.Context) (map[string]interface{}, error)
}

func GetCache() *Cache {
	return &Cache{
		data: sync.Map{},
		mu:   sync.RWMutex{},
	}
}

func (c *Cache) Get(ctx context.Context, key string) ([]*model.Ticket, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	value, ok := c.data.Load(key)
	if !ok {
		return nil, errors.New("key not found")
	}
	return value.([]*model.Ticket), nil
}

func (c *Cache) Set(ctx context.Context, key string, value []*model.Ticket, expiration time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data.Store(key, value)
	return nil
}

func (c *Cache) Del(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data.Delete(key)
	return nil
}

func (c *Cache) Len(ctx context.Context, key string) (int, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	value, ok := c.data.Load(key)
	if !ok {
		return 0, errors.New("key not found")
	}
	return len(value.([]*model.Ticket)), nil
}

func (c *Cache) Edit(ctx context.Context, key string, value *model.Ticket) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data.Store(key, value)
	return nil
}

// 获取缓存统计信息
func (c *Cache) GetStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 统计缓存中的键数量
	count := 0
	c.data.Range(func(key, value interface{}) bool {
		count++
		return true
	})

	stats["cache_count"] = count
	stats["cache_type"] = "local_memory"

	return stats, nil
}
