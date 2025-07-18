package model

import (
	"time"
)

type User struct {
	UserId       string    `json:"user_id" gorm:"column:user_id;primaryKey"`
	UserIdentity string    `json:"user_identity" gorm:"column:user_identity"`
	UserName     string    `json:"user_name" gorm:"column:user_name"`
	UserPwd      string    `json:"user_pwd" gorm:"column:user_pwd"`
	UserPhone    string    `json:"user_phone" gorm:"column:user_phone"`
	CreateTime   time.Time `json:"create_at" gorm:"column:create_at"`
	UpdateTime   time.Time `json:"update_at" gorm:"column:update_at"`
	DeleteTime   time.Time `json:"delete_at" gorm:"column:delete_at"`
}
