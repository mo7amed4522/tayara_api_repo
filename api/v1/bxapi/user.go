package bxapi

import (
	"github.com/gin-gonic/gin"
	bxapiReq "github.com/mo7amed4522/tayara_api/model/bxapi/request"
	"github.com/mo7amed4522/tayara_api/model/common/response"
	"github.com/mo7amed4522/tayara_api/utils"
)

type UserApi struct{}

// Register
// @Tags User
// @Summary Public end-user registration
// @Produce application/json
// @Param data body bxapiReq.UserRegisterReq true "user payload"
// @Success 200 {object} response.Response
// @Router /api/v1/user/register [post]
func (UserApi) Register(c *gin.Context) {
	var req bxapiReq.UserRegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	user, err := userService.UserServices.Register(req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(user, c)
}

// Login
// @Tags User
// @Summary User login (returns X-token)
// @Produce application/json
// @Param data body bxapiReq.UserLoginReq true "login"
// @Success 200 {object} response.Response
// @Router /api/v1/user/login [post]
func (UserApi) Login(c *gin.Context) {
	var req bxapiReq.UserLoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	res, err := userService.UserServices.Login(req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(res, c)
}

// FirebaseLogin
// @Tags User
// @Summary User login/register via Firebase (Google / Apple / Phone)
// @Produce application/json
// @Param data body bxapiReq.FirebaseLoginReq true "firebase id token"
// @Success 200 {object} response.Response
// @Router /api/v1/user/firebase-login [post]
func (UserApi) FirebaseLogin(c *gin.Context) {
	var req bxapiReq.FirebaseLoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	res, err := userService.UserServices.FirebaseLogin(c.Request.Context(), req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(res, c)
}

// GetProfile
// @Tags User
// @Summary User gets their own profile
// @Produce application/json
// @Success 200 {object} response.Response
// @Router /api/v1/user/profile [get]
func (UserApi) GetProfile(c *gin.Context) {
	userID := utils.GetBxUserID(c)
	user, err := userService.UserServices.GetProfile(userID)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(user, c)
}

// ChangePassword
// @Tags User
// @Summary User changes their own password
// @Produce application/json
// @Param data body bxapiReq.ChangeUserPasswordReq true "password change"
// @Success 200 {object} response.Response
// @Router /api/v1/user/change-password [post]
func (UserApi) ChangePassword(c *gin.Context) {
	userID := utils.GetBxUserID(c)
	var req bxapiReq.ChangeUserPasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := userService.UserServices.ChangePassword(userID, req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("password changed", c)
}

// ListUsers
// @Tags User
// @Summary Admin lists users (paginated, filterable by status)
// @Produce application/json
// @Param data query bxapiReq.UserPageReq true "page filter"
// @Success 200 {object} response.Response
// @Router /api/v1/user/admin [get]
func (UserApi) ListUsers(c *gin.Context) {
	var req bxapiReq.UserPageReq
	_ = c.ShouldBindQuery(&req)
	list, total, err := userService.UserServices.List(req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(response.PageResult{List: list, Total: total, Page: req.Page, PageSize: req.PageSize}, "ok", c)
}

// Block
// @Tags User
// @Summary Admin blocks/unblocks a user
// @Produce application/json
// @Param data body bxapiReq.UserBlockReq true "block payload"
// @Success 200 {object} response.Response
// @Router /api/v1/user/admin/block [post]
func (UserApi) Block(c *gin.Context) {
	adminID := utils.GetUserID(c)
	var req bxapiReq.UserBlockReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	user, err := userService.UserServices.Block(adminID, req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(user, c)
}
