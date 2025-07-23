package service

import (
	"12305/enum"
	"12305/model"
	"12305/mq/sender"
	"12305/repository"
	"12305/response"
	"12305/utils"
	"context"
	"errors"
	"fmt"
	"time"
)

type TicketService struct {
	ticketRepo   repository.TicketRepository
	redisRepo    repository.RedisRepository
	localRepo    repository.LocalRepository
	rabbitmqRepo sender.SenderStruct
}

type TicketSrv interface {
	ListByTicketTag(ctx context.Context, trainnumber interface{}) ([]*model.Ticket, error)
	// GetTotal(ctx context.Context, req *query.ListQuery) (int64, error)
	Get(ctx context.Context, ticket *model.Ticket) (*model.Ticket, error)
	GetByTicketTag(ctx context.Context, TicketTag []enum.TicketTag) ([]*model.Ticket, error)
	GetByTicketNumber(ctx context.Context, seat int) (*model.Ticket, error)
	// BuyTicket(ctx context.Context, ticket *model.Ticket, user response.User) (bool, error)
	//缓存穿透模式
	ListByTicketTagReadThrough(ctx context.Context, trainnumber interface{}) ([]*model.Ticket, error)
	BuyTicketWriteThrough(ctx context.Context, ticket *model.Ticket, user response.User) (bool, error)
	Exist(ctx context.Context, ticket *model.Ticket) (bool, error)
	Create(ctx context.Context, ticket *model.Ticket) (*model.Ticket, error)
	Edit(ctx context.Context, ticket *model.Ticket) (bool, error)
	Delete(ctx context.Context, ticket *model.Ticket) (bool, error)
	// 新增：缓存管理
	WarmUpCache(ctx context.Context) error
	GetCacheStats(ctx context.Context) (map[string]interface{}, error)
}

func (s *TicketService) ListByTicketTag(ctx context.Context, trainnumber interface{}) ([]*model.Ticket, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	//通过本地缓存获取所有车票

	//通过redis获取所有同车次票
	tickettag := trainnumber.(string)
	return s.redisRepo.GetByTicketTag(ctx, tickettag)
}

// func (s *TicketService) GetTotal(req *query.ListQuery) (int64, error) {
// 	return s.ticketRepo.GetTotal(req)
// }

func (s *TicketService) Get(ctx context.Context, ticket *model.Ticket) (*model.Ticket, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return s.ticketRepo.Get(ctx, ticket)
}

func (s *TicketService) Exist(ctx context.Context, ticket *model.Ticket) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	exist, err := s.ticketRepo.Exist(ctx, *ticket)
	if err != nil {
		fmt.Println("查询车票是否存在失败", err)
		return false, err
	}
	if !exist {
		fmt.Println("车票不存在")
		return false, nil
	}
	return exist, nil
}

// // 购买车票 - Write-Through模式
// func (s *TicketService) BuyTicket(ctx context.Context, ticket *model.Ticket, user response.User) (bool, error) {
// 	//获取分布式锁，防止同一张票被多个用户同时购买
// 	lockAcquired, err := s.redisRepo.AcquireTicketLock(ctx, ticket.TicketId, 10*time.Second)
// 	if err != nil {
// 		return false, fmt.Errorf("获取分布式锁失败: %v", err)
// 	}
// 	if !lockAcquired {
// 		return false, errors.New("票务繁忙，请稍后重试")
// 	}

// 	// 确保锁会被释放
// 	defer func() {
// 		if err := s.redisRepo.ReleaseTicketLock(ctx, ticket.TicketId); err != nil {
// 			fmt.Printf("释放分布式锁失败: %v\n", err)
// 		}
// 	}()

// 	// Write-Through模式：直接更新数据库和缓存
// 	err = s.ticketRepo.ExecuteTransaction(func(r *repository.TicketRepository) error {
// 		// 1. 获取当前票务信息（仅用于获取版本号和验证状态）
// 		currentTicket, err := r.Get(ctx, ticket)
// 		if err != nil {
// 			return fmt.Errorf("获取票务信息失败: %v", err)
// 		}

// 		// 2. 验证票务状态
// 		if currentTicket.TicketStatus != enum.TicketStatusNormal {
// 			return errors.New("票已售出或不可用")
// 		}

// 		// 3. Write-Through核心：使用乐观锁直接更新数据库
// 		success, err := r.UpdateTicketStatusWithOptimisticLock(ctx, ticket.TicketId, currentTicket.Version, enum.TicketStatusSold)
// 		if err != nil {
// 			return fmt.Errorf("更新票务状态失败: %v", err)
// 		}

// 		if !success {
// 			return errors.New("抢票失败，票已被其他用户购买")
// 		}

// 		// 4. Write-Through: 同步更新Redis缓存
// 		if err := s.redisRepo.SyncTicketToCache(ctx, currentTicket); err != nil {
// 			return fmt.Errorf("更新Redis缓存失败: %v", err)
// 		}

// 		// 5. 创建订单
// 		order := &model.Order{
// 			OrderId:     utils.GetUUID(),
// 			OrderStatus: enum.OrderStatusPending,
// 			TotalPrice:  currentTicket.TicketPrice,
// 			CreateTime:  time.Now(),
// 			UpdateTime:  time.Now(),
// 			User:        user,
// 			Ticket:      *currentTicket,
// 		}

// 		// 6. 发送订单到消息队列
// 		if err := s.rabbitmqRepo.SendOrder(ctx, *order); err != nil {
// 			return fmt.Errorf("发送订单到消息队列失败: %v", err)
// 		}

// 		return nil
// 	})

// 	if err != nil {
// 		fmt.Printf("抢票失败: %v\n", err)
// 		return false, err
// 	}

// 	// 7. 使本地缓存失效，强制重新加载
// 	if err := s.localRepo.InvalidateCache(ctx, string(ticket.TicketTag)); err != nil {
// 		fmt.Printf("使本地缓存失效失败: %v\n", err)
// 	}

// 	fmt.Println("抢票成功")
// 	return true, nil
// }

// 真正的Write-Through模式 - 将缓存视为主要数据存储
func (s *TicketService) BuyTicketWriteThrough(ctx context.Context, ticket *model.Ticket, user response.User) (bool, error) {
	//获取分布式锁，防止同一张票被多个用户同时购买
	lockAcquired, err := s.redisRepo.AcquireTicketLock(ctx, ticket.TicketId, 10*time.Second)
	if err != nil {
		return false, fmt.Errorf("获取分布式锁失败: %v", err)
	}
	if !lockAcquired {
		return false, errors.New("票务繁忙，请稍后重试")
	}

	// 确保锁会被释放
	defer func() {
		if err := s.redisRepo.ReleaseTicketLock(ctx, ticket.TicketId); err != nil {
			fmt.Printf("释放分布式锁失败: %v\n", err)
		}
	}()

	// Write-Through模式：同时写入缓存和数据库
	err = s.ticketRepo.ExecuteTransaction(func(r *repository.TicketRepository) error {
		// 获取当前票务信息
		currentTicket, err := r.Get(ctx, ticket)
		if err != nil {
			return fmt.Errorf("获取票务信息失败: %v", err)
		}

		// 验证票务状态
		if currentTicket.TicketStatus != enum.TicketStatusNormal {
			return errors.New("票已售出或不可用")
		}

		// 更新数据库
		success, err := r.UpdateTicketStatusWithOptimisticLock(ctx, ticket.TicketId, currentTicket.Version, enum.TicketStatusSold)
		if err != nil {
			return fmt.Errorf("更新票务状态失败: %v", err)
		}

		if !success {
			return errors.New("抢票失败，票已被其他用户购买")
		}

		// Write-Through核心：同时更新缓存
		// 更新票务状态
		currentTicket.TicketStatus = enum.TicketStatusSold
		currentTicket.Version++
		currentTicket.UpdateTime = time.Now()

		// 同步更新Redis缓存
		if err := s.redisRepo.SyncTicketToCache(ctx, currentTicket); err != nil {
			return fmt.Errorf("Write-Through: 更新Redis缓存失败: %v", err)
		}

		// 创建订单
		order := &model.Order{
			OrderId:     utils.GetUUID(),
			OrderStatus: enum.OrderStatusPending,
			TotalPrice:  currentTicket.TicketPrice,
			CreateTime:  time.Now(),
			UpdateTime:  time.Now(),
			User:        user,
			Ticket:      *currentTicket,
		}

		// 发送订单到消息队列
		if err := s.rabbitmqRepo.SendOrder(ctx, *order); err != nil {
			return fmt.Errorf("发送订单到消息队列失败: %v", err)
		}

		fmt.Printf("Write-Through: 已同时更新数据库和缓存\n")
		return nil
	})

	if err != nil {
		fmt.Printf("抢票失败: %v\n", err)
		return false, err
	}

	// 使本地缓存失效，强制重新加载
	if err := s.localRepo.InvalidateCache(ctx, string(ticket.TicketTag)); err != nil {
		fmt.Printf("使本地缓存失效失败: %v\n", err)
	}

	fmt.Println("抢票成功")
	return true, nil
}

func (s *TicketService) Create(ctx context.Context, ticket *model.Ticket) (*model.Ticket, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	exist, err := s.ticketRepo.Exist(ctx, *ticket)
	if err != nil {
		fmt.Println("查询车票是否存在失败", err)
		return nil, err
	}
	if exist {
		fmt.Println("车票已存在")
		return nil, nil
	}
	Ticket, err := s.ticketRepo.Get(ctx, ticket)
	if err != nil {
		fmt.Println("获取车票失败", err)
		return nil, err
	}
	if Ticket.TicketId == "" {
		Ticket.TicketId = utils.GetUUID()
	}
	Ticket.CreateTime = time.Now()
	Ticket.UpdateTime = time.Now()
	Ticket.TicketTag = ticket.TicketTag
	Ticket.TicketNumber = ticket.TicketNumber
	Ticket.TicketPrice = ticket.TicketPrice
	if Ticket.TicketPrice == 0 {
		Ticket.TicketPrice = 999.999
	}
	return s.ticketRepo.CreateTicket(ctx, Ticket)
}

func (s *TicketService) Edit(ctx context.Context, ticket *model.Ticket) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	exist, err := s.ticketRepo.Exist(ctx, *ticket)
	if err != nil {
		fmt.Println("查询车票是否存在失败", err)
		return false, err
	}
	if !exist {
		fmt.Println("车票不存在")
		return false, nil
	}
	Ticket, err := s.ticketRepo.Get(ctx, ticket)
	if err != nil {
		fmt.Println("获取车票失败", err)
		return false, err
	}
	Ticket.TicketTag = ticket.TicketTag
	Ticket.TicketNumber = ticket.TicketNumber
	Ticket.TicketPrice = ticket.TicketPrice
	Ticket.UpdateTime = time.Now()
	return s.ticketRepo.Edit(ctx, Ticket)
}

func (s *TicketService) Delete(ctx context.Context, ticket *model.Ticket) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	exist, err := s.ticketRepo.Exist(ctx, *ticket)
	if err != nil {
		fmt.Println("查询车票是否存在失败", err)
		return false, err
	}
	if !exist {
		fmt.Println("车票不存在")
		return false, nil
	}
	return s.ticketRepo.Delete(ctx, ticket)
}

// 缓存预热 - 系统启动时预加载热门车次数据
func (s *TicketService) WarmUpCache(ctx context.Context) error {
	// 定义热门车次标签
	popularTags := []enum.TicketTag{
		"G101", "G102", "G103", "G104", "G105",
		"D101", "D102", "D103", "D104", "D105",
		"K101", "K102", "K103", "K104", "K105",
	}

	fmt.Println("开始缓存预热...")

	for _, tag := range popularTags {
		// 从数据库获取该车次的所有票务信息
		tickets, err := s.ticketRepo.GetByTicketTag(ctx, []enum.TicketTag{tag})
		if err != nil {
			fmt.Printf("预热车次 %s 失败: %v\n", tag, err)
			continue
		}

		// 批量更新Redis缓存
		for _, ticket := range tickets {
			if err := s.redisRepo.SyncTicketToCache(ctx, ticket); err != nil {
				fmt.Printf("预热票务 %s 到Redis失败: %v\n", ticket.TicketId, err)
				continue
			}
		}

		// 预热本地缓存
		if err := s.localRepo.RefreshCache(ctx, string(tag)); err != nil {
			fmt.Printf("预热车次 %s 到本地缓存失败: %v\n", tag, err)
		}

		fmt.Printf("车次 %s 预热完成，共 %d 张票\n", tag, len(tickets))
	}

	fmt.Println("缓存预热完成")
	return nil
}

// 获取缓存统计信息
func (s *TicketService) GetCacheStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 获取Redis缓存统计
	redisStats, err := s.redisRepo.GetCacheStats(ctx)
	if err != nil {
		fmt.Printf("获取Redis缓存统计失败: %v\n", err)
	} else {
		stats["redis"] = redisStats
	}

	// 获取本地缓存统计
	localStats, err := s.localRepo.GetCacheStats(ctx)
	if err != nil {
		fmt.Printf("获取本地缓存统计失败: %v\n", err)
	} else {
		stats["local"] = localStats
	}

	return stats, nil
}

// 实现Read-Through模式 - 将缓存视为主要数据存储（防击穿、防穿透、防雪崩版）
func (s *TicketService) ListByTicketTagReadThrough(ctx context.Context, trainnumber interface{}) ([]*model.Ticket, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	tickettag := trainnumber.(string)

	// 1. 首先尝试从Redis缓存读取（Read-Through的核心）
	tickets, err := s.redisRepo.GetByTicketTag(ctx, tickettag)
	if err == nil && len(tickets) > 0 {
		fmt.Printf("Read-Through: 从Redis缓存获取到 %d 张票\n", len(tickets))
		return tickets, nil
	}

	// 2. 防击穿：使用互斥锁保护数据库访问
	protector := utils.GetCacheProtector()
	mutex := protector.GetMutex(tickettag)
	mutex.Lock()
	defer mutex.Unlock()

	// 3. 双重检查：再次尝试从缓存读取
	tickets, err = s.redisRepo.GetByTicketTag(ctx, tickettag)
	if err == nil && len(tickets) > 0 {
		fmt.Printf("Read-Through: 双重检查从Redis缓存获取到 %d 张票\n", len(tickets))
		return tickets, nil
	}

	// 4. 缓存未命中，从数据库读取
	fmt.Printf("Read-Through: Redis缓存未命中，从数据库加载车次 %s 的票务信息\n", tickettag)
	tickets, err = s.ticketRepo.GetByTicketTag(ctx, []enum.TicketTag{enum.TicketTag(tickettag)})
	if err != nil {
		return nil, fmt.Errorf("从数据库加载票务信息失败: %v", err)
	}

	// 5. 将数据写入缓存（Read-Through的关键）
	if len(tickets) > 0 {
		// 批量更新Redis缓存
		for _, ticket := range tickets {
			if err := s.redisRepo.SyncTicketToCache(ctx, ticket); err != nil {
				fmt.Printf("Read-Through: 更新Redis缓存失败: %v\n", err)
			}
		}

		// 更新本地缓存
		if err := s.localRepo.RefreshCache(ctx, tickettag); err != nil {
			fmt.Printf("Read-Through: 更新本地缓存失败: %v\n", err)
		}

		fmt.Printf("Read-Through: 已将 %d 张票加载到缓存\n", len(tickets))
	} else {
		// 防穿透：缓存空结果
		protector.CacheNull(tickettag, 5*time.Minute)
	}

	return tickets, nil
}
