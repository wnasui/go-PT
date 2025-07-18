package repository

import (
	"12305/model"
	"12305/query"
	"12305/utils"
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type UserRepository struct {
	DB *gorm.DB
}

type UserRepoInterface interface {
	List(ctx context.Context, req *query.ListQuery) ([]*model.User, error)
	GetTotal(ctx context.Context, req *query.ListQuery) (int64, error)
	Get(ctx context.Context, user *model.User) (*model.User, error)
	GetByUserIdentity(ctx context.Context, userIdentity string) (*model.User, error)
	GetByUserPhone(ctx context.Context, userPhone string) (*model.User, error)
	Exist(ctx context.Context, user *model.User) (bool, error)
	ExistByUserIdentity(ctx context.Context, userIdentity string) (bool, error)
	ExistByUserPhone(ctx context.Context, phone string) (bool, error)
	CreateUser(ctx context.Context, user *model.User) error
	Edit(ctx context.Context, user *model.User) (bool, error)
	Delete(ctx context.Context, user *model.User) (bool, error)
}

func (repo *UserRepository) List(ctx context.Context, req *query.ListQuery) (users []*model.User, err error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	fmt.Println(req)
	db := repo.DB
	limit, offset := utils.GetLimitAndOffset(req.Page, req.PageSize)
	err = db.Order("id desc").Limit(limit).Offset(offset).Find(&users).Error
	return users, err
}

func (repo *UserRepository) GetTotal(ctx context.Context, req *query.ListQuery) (total int64, err error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	var users []*model.User
	db := repo.DB

	err = db.Find(&users).Count(&total).Error
	if err != nil {
		return 0, err
	}
	return total, nil
}

func (repo *UserRepository) Get(ctx context.Context, user *model.User) (*model.User, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	db := repo.DB
	User := model.User{}
	err := db.Where("user_id=?", user.UserId).First(&User).Error
	if err != nil {
		return nil, err
	}
	return &User, nil
}

func (repo *UserRepository) GetByUserIdentity(ctx context.Context, userIdentity string) (*model.User, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	db := repo.DB
	User := model.User{}
	err := db.Where("user_identity=?", userIdentity).First(&User).Error
	if err != nil {
		return nil, err
	}
	return &User, nil
}

func (repo *UserRepository) GetByUserPhone(ctx context.Context, userPhone string) (*model.User, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	db := repo.DB
	User := model.User{}
	err := db.Where("user_phone=?", userPhone).First(&User).Error
	if err != nil {
		return nil, err
	}
	return &User, nil
}

func (repo *UserRepository) Exist(ctx context.Context, user *model.User) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	db := repo.DB

	err := db.Where("user_id=?", user.UserId).First(&user).Error
	if user.UserId == "" {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (repo *UserRepository) ExistByUserIdentity(ctx context.Context, userIdentity string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	db := repo.DB
	var user model.User

	err := db.Where("user_identity=?", userIdentity).First(&user).Error
	if err != nil {
		return false, err
	}
	return true, nil
}

func (repo *UserRepository) ExistByUserPhone(ctx context.Context, phone string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	db := repo.DB
	var user model.User
	err := db.Where("user_phone=?", phone).First(&user).Error
	if err != nil {
		return false, err
	}
	return true, nil
}

func (repo *UserRepository) CreateUser(ctx context.Context, user *model.User) (*model.User, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	db := repo.DB
	exist, err := repo.Exist(ctx, user)
	if err != nil {
		return nil, err
	}
	if exist {
		fmt.Println("用户已存在")
		return nil, nil
	}

	err = db.Create(user).Error
	if err != nil {
		fmt.Println("用户注册失败", err)
		return nil, err
	}
	return user, nil
}

func (repo *UserRepository) Edit(ctx context.Context, user *model.User) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	db := repo.DB
	err := db.Model(&user).Where("user_id=?", user.UserId).Updates(map[string]interface{}{
		"user_name":   user.UserName,
		"user_pwd":    user.UserPwd,
		"user_phone":  user.UserPhone,
		"update_time": time.Now(),
	}).Error
	if err != nil {
		return false, err
	}
	return true, nil
}

func (repo *UserRepository) Delete(ctx context.Context, user *model.User) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	db := repo.DB
	err := db.Model(&user).Where("user_id=?", user.UserId).Delete(&user).Error
	if err != nil {
		return false, err
	}
	return true, nil
}
