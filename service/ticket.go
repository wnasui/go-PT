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
	BuyTicket(ctx context.Context, ticket *model.Ticket, user response.User) (bool, error)
	Exist(ctx context.Context, ticket *model.Ticket) (bool, error)
	Create(ctx context.Context, ticket *model.Ticket) (*model.Ticket, error)
	Edit(ctx context.Context, ticket *model.Ticket) (bool, error)
	Delete(ctx context.Context, ticket *model.Ticket) (bool, error)
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

// BuyTicket 购买车票 - 使用分布式锁+乐观锁+缓存一致性管理
func (s *TicketService) BuyTicket(ctx context.Context, ticket *model.Ticket, user response.User) (bool, error) {
	// 1. 获取分布式锁，防止同一张票被多个用户同时购买
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

	// 2. 在事务中执行购票逻辑
	err = s.ticketRepo.ExecuteTransaction(func(r *repository.TicketRepository) error {
		// 3. 查询数据库中的票务状态
		currentTicket, err := r.Get(ctx, ticket)
		if err != nil {
			return fmt.Errorf("查询票务信息失败: %v", err)
		}

		// 4. 检查票务状态
		if currentTicket.TicketStatus != enum.TicketStatusNormal {
			return errors.New("票已售出或不可用")
		}

		// 5. 使用乐观锁更新票务状态
		success, err := r.UpdateTicketStatusWithOptimisticLock(ctx, ticket.TicketId, currentTicket.Version, enum.TicketStatusSold)
		if err != nil {
			return fmt.Errorf("更新票务状态失败: %v", err)
		}

		if !success {
			return errors.New("抢票失败，票已被其他用户购买")
		}

		// 6. 创建订单
		order := &model.Order{
			OrderId:     utils.GetUUID(),
			OrderStatus: enum.OrderStatusPending,
			TotalPrice:  currentTicket.TicketPrice,
			CreateTime:  time.Now(),
			UpdateTime:  time.Now(),
			User:        user,
			Ticket:      *currentTicket,
		}

		// 7. 发送订单到消息队列
		if err := s.rabbitmqRepo.SendOrder(ctx, *order); err != nil {
			return fmt.Errorf("发送订单到消息队列失败: %v", err)
		}

		// 8. 更新缓存 - 异步处理，不影响主流程
		go func() {
			// 更新Redis缓存
			if err := s.redisRepo.SyncTicketToCache(ctx, currentTicket); err != nil {
				fmt.Printf("更新Redis缓存失败: %v\n", err)
			}

			// 使本地缓存失效，强制重新加载
			if err := s.localRepo.InvalidateCache(ctx, string(currentTicket.TicketTag)); err != nil {
				fmt.Printf("使本地缓存失效失败: %v\n", err)
			}
		}()

		return nil
	})

	if err != nil {
		fmt.Printf("抢票失败: %v\n", err)
		return false, err
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
