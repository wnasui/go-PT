package utils

import (
	"context"
	"crypto/md5"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// 缓存保护器 - 防止击穿、穿透、雪崩
type CacheProtector struct {
	mutexMap  sync.Map
	nullCache sync.Map
	hotKeys   sync.Map
}

var protector = &CacheProtector{}

func GetCacheProtector() *CacheProtector {
	return protector
}

// 防击穿：获取互斥锁
func (cp *CacheProtector) GetMutex(key string) *sync.Mutex {
	hashKey := fmt.Sprintf("%x", md5.Sum([]byte(key)))

	mutex, _ := cp.mutexMap.LoadOrStore(hashKey, &sync.Mutex{})
	return mutex.(*sync.Mutex)
}

// 防穿透：检查是否为空值缓存
func (cp *CacheProtector) IsNullCached(key string) bool {
	_, exists := cp.nullCache.Load(key)
	return exists
}

// 防穿透：缓存空值
func (cp *CacheProtector) CacheNull(key string, duration time.Duration) {
	cp.nullCache.Store(key, time.Now().Add(duration))

	go func() {
		time.Sleep(duration)
		cp.nullCache.Delete(key)
	}()
}

// 防雪崩：生成随机过期时间
func (cp *CacheProtector) GetRandomExpiration(baseDuration time.Duration) time.Duration {
	randomFactor := 1 + rand.Float64()*0.3
	return time.Duration(float64(baseDuration) * randomFactor)
}

// 标记热点数据
func (cp *CacheProtector) MarkHotKey(key string) {
	cp.hotKeys.Store(key, time.Now())
}

func (cp *CacheProtector) IsHotKey(key string) bool {
	_, exists := cp.hotKeys.Load(key)
	return exists
}

func (cp *CacheProtector) GetHotKeysStats() map[string]interface{} {
	stats := make(map[string]interface{})
	count := 0

	cp.hotKeys.Range(func(key, value interface{}) bool {
		count++
		return true
	})

	stats["hot_keys_count"] = count
	return stats
}

type BloomFilter struct {
	bitmap []bool
	size   int
	hashes int
}

func NewBloomFilter(size, hashes int) *BloomFilter {
	return &BloomFilter{
		bitmap: make([]bool, size),
		size:   size,
		hashes: hashes,
	}
}

func (bf *BloomFilter) Add(key string) {
	for i := 0; i < bf.hashes; i++ {
		hash := bf.hash(key, i)
		bf.bitmap[hash%bf.size] = true
	}
}

func (bf *BloomFilter) MayContain(key string) bool {
	for i := 0; i < bf.hashes; i++ {
		hash := bf.hash(key, i)
		if !bf.bitmap[hash%bf.size] {
			return false
		}
	}
	return true
}

// 哈希函数
func (bf *BloomFilter) hash(key string, seed int) int {
	hash := 0
	for _, char := range key {
		hash = (hash*31 + int(char) + seed) % bf.size
	}
	return hash
}

// 限流器
type RateLimiter struct {
	tokens    chan struct{}
	rate      time.Duration
	burst     int
	lastReset time.Time
	mu        sync.Mutex
}

// 创建限流器
func NewRateLimiter(rate time.Duration, burst int) *RateLimiter {
	rl := &RateLimiter{
		tokens: make(chan struct{}, burst),
		rate:   rate,
		burst:  burst,
	}

	// 初始化令牌桶
	for i := 0; i < burst; i++ {
		rl.tokens <- struct{}{}
	}

	// 启动令牌补充
	go rl.refillTokens()

	return rl
}

// 补充令牌
func (rl *RateLimiter) refillTokens() {
	ticker := time.NewTicker(rl.rate)
	defer ticker.Stop()

	for range ticker.C {
		select {
		case rl.tokens <- struct{}{}:
		default:
			// 令牌桶已满
		}
	}
}

// 尝试获取令牌
func (rl *RateLimiter) Allow() bool {
	select {
	case <-rl.tokens:
		return true
	default:
		return false
	}
}

// 等待获取令牌
func (rl *RateLimiter) Wait(ctx context.Context) error {
	select {
	case <-rl.tokens:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
