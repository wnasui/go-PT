package handler

import (
	"12305/enum"
	"12305/model"
	"12305/response"
	"12305/service"
	"12305/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	UserService service.UserSrv
}

func (h *UserHandler) GetEntity(user model.User) response.User {
	return response.User{
		ID:        utils.GetUUID(),
		Key:       utils.GetUUID(),
		UserId:    user.UserId,
		UserPhone: user.UserPhone,
		UserName:  user.UserName,
		UserPwd:   user.UserPwd,
		CreatedAt: user.CreateTime,
		UpdatedAt: user.UpdateTime,
	}
}

func (h *UserHandler) UserInfoHandler(c *gin.Context) {
	entity := response.Entity{
		Code:  int(enum.OperateOK),
		Msg:   enum.OperateOK.String(),
		Total: 0,
		Data:  nil,
	}

	userId := c.Param("user_id")
	if userId == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"entity": entity})
		return
	}
	user := model.User{
		UserId: userId,
	}
	result, err := h.UserService.Get(c, &user)
	if err != nil {
		entity.Code = int(enum.OperateFailed)
		entity.Msg = enum.OperateFailed.String()
		c.JSON(http.StatusInternalServerError, gin.H{"entity": entity})
		return
	}

	r := h.GetEntity(*result)
	entity.Data = r
	entity.Msg = "success"
	c.JSON(http.StatusOK, gin.H{"entity": entity})
}

// func (h *UserHandler) UserListHandler(c *gin.Context) {
// 	var q query.ListQuery
// 	entity := response.Entity{
// 		Code:      int(enum.OperateOK),
// 		Msg:       enum.OperateOK.String(),
// 		Total:     0,
// 		TotalPage: 1,
// 		Data:      nil,
// 	}
// 	if err := c.ShouldBindQuery(&q); err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"entity": entity})
// 		return
// 	}

// 	list, err := h.userService.List(&q)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"entity": entity})
// 		return
// 	}
// 	total, err := h.userService.GetTotal(&q)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"entity": entity})
// 		return
// 	}
// 	if q.PageSize == 0 {
// 		q.PageSize = config.PageSize
// 	}
// 	ret := int(total) / q.PageSize
// 	ret2 := int(total) % q.PageSize
// 	if ret2 > 0 {
// 		ret++
// 	}
// 	var newList []response.User
// 	for _, v := range list {
// 		newList = append(newList, h.GetEntity(*v))
// 	}
// 	entity.Total = int(total)
// 	entity.TotalPage = ret
// 	entity.Data = newList
// 	entity.Msg = "success"
// 	c.JSON(http.StatusOK, gin.H{"entity": entity})

// }

func (h *UserHandler) UserCreateHandler(c *gin.Context) {
	entity := response.Entity{
		Code:  int(enum.OperateOK),
		Msg:   enum.OperateOK.String(),
		Total: 0,
		Data:  nil,
	}

	var user model.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"entity": entity})
		return
	}

	_, err := h.UserService.Create(c, &user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"entity": entity})
		return
	}

	entity.Msg = "success"
	c.JSON(http.StatusOK, gin.H{"entity": entity})
}

func (h *UserHandler) UserEditHandler(c *gin.Context) {
	entity := response.Entity{
		Code:  int(enum.OperateOK),
		Msg:   enum.OperateOK.String(),
		Total: 0,
		Data:  nil,
	}

	var user model.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"entity": entity})
		return
	}
	b, err := h.UserService.Edit(c, &user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"entity": entity})
		return
	}
	if !b {
		entity.Code = int(enum.OperateFailed)
		entity.Msg = enum.OperateFailed.String()
		c.JSON(http.StatusInternalServerError, gin.H{"entity": entity})
		return
	}

	entity.Msg = "success"
	c.JSON(http.StatusOK, gin.H{"entity": entity})
}

func (h *UserHandler) UserDeleteHandler(c *gin.Context) {
	entity := response.Entity{
		Code:  int(enum.OperateOK),
		Msg:   enum.OperateOK.String(),
		Total: 0,
		Data:  nil,
	}
	userId := c.Param("user_id")
	if userId == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"entity": entity})
		return
	}
	user := model.User{
		UserId: userId,
	}
	_, err := h.UserService.Delete(c, &user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"entity": entity})
		return
	}

	entity.Msg = "success"
	c.JSON(http.StatusOK, gin.H{"entity": entity})
}

func (h *UserHandler) UserLoginHandler(c *gin.Context) {
	entity := response.Entity{
		Code:  int(enum.OperateOK),
		Msg:   enum.OperateOK.String(),
		Total: 0,
		Data:  nil,
	}
	var user model.User
	err := c.ShouldBindJSON(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"entity": entity})
		return
	}
	result, err := h.UserService.Login(c, &user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"entity": entity})
		return
	}
	if result == nil {
		entity.Code = int(enum.OperateFailed)
		entity.Msg = enum.OperateFailed.String()
		c.JSON(http.StatusInternalServerError, gin.H{"entity": entity})
		return
	}
	entity.Data = h.GetEntity(*result)
	entity.Msg = "success"
	c.JSON(http.StatusOK, gin.H{"entity": entity})
}
