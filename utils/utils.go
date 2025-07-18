package utils

import (
	"12305/model"
	"crypto/md5"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// 分页
func GetLimitAndOffset(page, pageSize int) (limit, offset int) {
	if pageSize > 0 {
		limit = pageSize
	} else {
		limit = 10
	}
	if page > 0 {
		offset = (page - 1) * limit
	} else {
		offset = 0
	}
	return limit, offset
}

// 时间
const TimeLayout = "2006-01-02 15:04:05"

var (
	local, _ = time.LoadLocation("Asia/beijing")
)

func GetTime() string {
	now := time.Now().In(local)
	return now.Format(TimeLayout)
}

func TimeFormat(s string) string {
	result, err := time.ParseInLocation(TimeLayout, s, local)
	if err != nil {
		panic(err)
	}
	return result.In(local).Format(TimeLayout)
}

// md5加密
func Md5(s string) string {
	w := md5.New()
	io.WriteString(w, s)
	Md5str := fmt.Sprintf("%x", w.Sum(nil))
	return Md5str
}

// UUID
func GetUUID() string {
	return uuid.New().String()
}

// 系统启动时批量加载本地缓存
var localCache sync.Map

var DB *gorm.DB

func QueryAllTicketsForToday() ([]*model.Ticket, error) {
	var tickets []*model.Ticket
	err := DB.Where("date = ?", time.Now().Format("2006-01-02")).Find(&tickets).Error
	if err != nil {
		return nil, err
	}
	return tickets, nil
}

func InitLocalCache() {
	tickets, err := QueryAllTicketsForToday()
	if err != nil {
		panic(err)
	}
	for _, ticket := range tickets {
		localCache.Store(ticket.TicketTag, ticket)
	}
}
