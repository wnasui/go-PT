package service

import (
	"12305/config"
	"12305/model"
	"12305/repository"
	"12305/utils"
	"context"
	"fmt"
	"time"

	"12305/query"
)

type UserService struct {
	userRepo repository.UserRepository
}

type UserSrv interface {
	List(ctx context.Context, req *query.ListQuery) ([]*model.User, error)
	GetTotal(ctx context.Context, req *query.ListQuery) (int, error)
	Get(ctx context.Context, user *model.User) (*model.User, error)
	GetByUserIdentity(ctx context.Context, userIdentity string) (*model.User, error)
	Exist(ctx context.Context, user *model.User) (bool, error)
	Create(ctx context.Context, user *model.User) (*model.User, error)
	Login(ctx context.Context, user *model.User) (*model.User, error)
	Edit(ctx context.Context, user *model.User) (bool, error)
	Delete(ctx context.Context, user *model.User) (*model.User, error)
}

func (s *UserService) List(ctx context.Context, req *query.ListQuery) ([]*model.User, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if req.PageSize <= 1 {
		req.PageSize = config.PageSize
	}
	return s.userRepo.List(ctx, req)
}

func (s *UserService) GetTotal(ctx context.Context, req *query.ListQuery) (int64, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	return s.userRepo.GetTotal(ctx, req)
}

func (s *UserService) Get(ctx context.Context, user *model.User) (*model.User, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return s.userRepo.Get(ctx, user)
}

func (s *UserService) GetByUserIdentity(ctx context.Context, userIdentity string) (*model.User, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	User, err := s.userRepo.GetByUserIdentity(ctx, userIdentity)
	if err != nil {
		fmt.Println("查询用户失败", err)
		return nil, err
	}
	if User == nil {
		fmt.Println("用户不存在")
		return nil, nil
	}
	return User, nil
}

func (s *UserService) Exist(ctx context.Context, user *model.User) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	return s.userRepo.Exist(ctx, user)
}

func (s *UserService) Create(ctx context.Context, user *model.User) (*model.User, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	result, err := s.userRepo.Exist(ctx, user)
	if err != nil {
		fmt.Println("查询用户是否存在失败", err)
		return nil, err
	}
	if result {
		fmt.Println("用户已存在")
		return nil, nil
	}
	user.UserId = utils.GetUUID()
	return s.userRepo.CreateUser(ctx, user)
}

func (s *UserService) Login(ctx context.Context, user *model.User) (*model.User, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	User, err := s.userRepo.GetByUserPhone(ctx, user.UserPhone)
	if err != nil {
		fmt.Println("用户账号不存在", err)
		return nil, err
	}
	if User.UserPwd != user.UserPwd {
		fmt.Println("密码错误")
		return nil, nil
	}
	return User, nil
}

func (s *UserService) Edit(ctx context.Context, user *model.User) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	result, err := s.userRepo.Exist(ctx, user)
	if err != nil {
		fmt.Println("查询用户是否存在失败", err)
		return false, err
	}
	if !result {
		fmt.Println("用户不存在")
		return false, nil
	}
	exist, err := s.userRepo.Get(ctx, user)
	if err != nil {
		fmt.Println("查询用户失败", err)
		return false, err
	}
	if exist == nil {
		fmt.Println("用户不存在")
		return false, nil
	}
	exist.UserName = user.UserName
	exist.UserPhone = user.UserPhone
	exist.UpdateTime = time.Now()
	exist.UserPwd = user.UserPwd

	return s.userRepo.Edit(ctx, exist)
}

func (s *UserService) Delete(ctx context.Context, user *model.User) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	result, err := s.userRepo.Exist(ctx, user)
	if err != nil {
		fmt.Println("查询用户是否存在失败", err)
		return false, err
	}
	if !result {
		fmt.Println("用户不存在")
		return false, nil
	}
	return s.userRepo.Delete(ctx, user)
}
