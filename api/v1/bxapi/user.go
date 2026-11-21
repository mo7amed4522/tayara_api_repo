package bxapi

import (
	"github.com/gin-gonic/gin"
	bxapiReq "github.com/mo7amed4522/tayara_api/model/bxapi/request"
	bxapiResp "github.com/mo7amed4522/tayara_api/model/bxapi/response"
	"github.com/mo7amed4522/tayara_api/model/common/response"
	"github.com/mo7amed4522/tayara_api/utils"
)

type UserApi struct{}

// Register
//
//	@Tags		User
//	@Summary	Public end-user registration
//	@Produce	application/json
//	@Param		data	body		bxapiReq.UserRegisterReq	true	"user payload"
//	@Success	200		{object}	response.Response
//	@Router		/api/v1/user/register [post]
func (UserApi) Register(c *gin.Context) {
	var req bxapiReq.UserRegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	user, err := bxapiSvc.UserServices.Register(req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(user, c)
}

// Login
//
//	@Tags		User
//	@Summary	User login (returns X-token)
//	@Produce	application/json
//	@Param		data	body		bxapiReq.UserLoginReq	true	"login"
//	@Success	200		{object}	response.Response
//	@Router		/api/v1/user/login [post]
func (UserApi) Login(c *gin.Context) {
	var req bxapiReq.UserLoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	res, err := bxapiSvc.UserServices.Login(req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(res, c)
}

// FirebaseLogin
//
//	@Tags		User
//	@Summary	User login/register via Firebase (Google / Apple / Phone)
//	@Produce	application/json
//	@Param		data	body		bxapiReq.FirebaseLoginReq	true	"firebase id token"
//	@Success	200		{object}	response.Response
//	@Router		/api/v1/user/firebase-login [post]
func (UserApi) FirebaseLogin(c *gin.Context) {
	var req bxapiReq.FirebaseLoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	res, err := bxapiSvc.UserServices.FirebaseLogin(c.Request.Context(), req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(res, c)
}

// GetProfile
//
//	@Tags		User
//	@Summary	User gets their own profile with trip/favorite counts
//	@Produce	application/json
//	@Success	200	{object}	response.Response
//	@Router		/api/v1/user/profile [get]
func (UserApi) GetProfile(c *gin.Context) {
	userID := utils.GetBxUserID(c)
	data, err := bxapiSvc.UserServices.GetProfile(userID)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(data, c)
}

// UpdateProfile
//
//	@Tags		User
//	@Summary	User updates their own profile
//	@Accept		multipart/form-data
//	@Produce	application/json
//	@Param		fullName			formData	string	false	"Full name"
//	@Param		nickName			formData	string	false	"Nickname"
//	@Param		email				formData	string	false	"Email"
//	@Param		phoneCountryCode	formData	string	false	"Phone country code"
//	@Param		phone				formData	string	false	"Phone number"
//	@Param		nationality			formData	string	false	"Nationality"
//	@Param		dateOfBirth			formData	string	false	"Date of birth (RFC3339)"
//	@Param		headerImg			formData	file	false	"Profile image"
//	@Success	200					{object}	response.Response
//	@Router		/api/v1/user/profile [put]
func (UserApi) UpdateProfile(c *gin.Context) {
	userID := utils.GetBxUserID(c)

	// Get optional file upload
	file, _ := c.FormFile("headerImg")

	var req bxapiReq.UserUpdateProfileReq
	if err := c.ShouldBind(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	user, err := bxapiSvc.UserServices.UpdateProfile(userID, req, file)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(bxapiResp.BxUpdateUserProfileData{User: *user}, c)
}

// DeleteAccount
//
//	@Tags		User
//	@Summary	User deletes their own account
//	@Produce	application/json
//	@Success	200	{object}	response.Response
//	@Router		/api/v1/user/account [delete]
func (UserApi) DeleteAccount(c *gin.Context) {
	userID := utils.GetBxUserID(c)
	if err := bxapiSvc.UserServices.DeleteAccount(userID); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("account deleted", c)
}

// ChangePassword
//
//	@Tags		User
//	@Summary	User changes their own password
//	@Produce	application/json
//	@Param		data	body		bxapiReq.ChangeUserPasswordReq	true	"password change"
//	@Success	200		{object}	response.Response
//	@Router		/api/v1/user/change-password [post]
func (UserApi) ChangePassword(c *gin.Context) {
	userID := utils.GetBxUserID(c)
	var req bxapiReq.ChangeUserPasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := bxapiSvc.UserServices.ChangePassword(userID, req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("password changed", c)
}

// ListUsers
//
//	@Tags		User
//	@Summary	Admin lists users (paginated, filterable by status)
//	@Produce	application/json
//	@Param		data	query		bxapiReq.UserPageReq	true	"page filter"
//	@Success	200		{object}	response.Response
//	@Router		/api/v1/user/admin [get]
func (UserApi) ListUsers(c *gin.Context) {
	var req bxapiReq.UserPageReq
	_ = c.ShouldBindQuery(&req)
	list, total, totalPage, err := bxapiSvc.UserServices.List(req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(response.PageResult{List: list, Total: total, TotalPage: totalPage, Page: req.Page, PageSize: req.PageSize}, "ok", c)
}

// Block
//
//	@Tags		User
//	@Summary	Admin blocks/unblocks a user
//	@Produce	application/json
//	@Param		data	body		bxapiReq.UserBlockReq	true	"block payload"
//	@Success	200		{object}	response.Response
//	@Router		/api/v1/user/admin/block [post]
func (UserApi) Block(c *gin.Context) {
	adminID := utils.GetUserID(c)
	var req bxapiReq.UserBlockReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	user, err := bxapiSvc.UserServices.Block(adminID, req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(user, c)
}
