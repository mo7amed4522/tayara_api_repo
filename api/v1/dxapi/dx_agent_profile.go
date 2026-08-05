package dxapi

import (
	"mime/multipart"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/middleware"
	"github.com/mo7amed4522/tayara_api/model/common/response"
	"github.com/mo7amed4522/tayara_api/model/dxapi"
	"github.com/mo7amed4522/tayara_api/model/dxapi/request"
	"github.com/mo7amed4522/tayara_api/model/system"
	"github.com/mo7amed4522/tayara_api/utils"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type DxAgentProfileApi struct{}

// Register
// @Tags SystemAuth
// @Summary Register system Agent
// @Produce application/json
// @Param data body request.DxCreateAgentProfileRequest true "register"
// @Success 200 {object} response.Response
// @Router /api/v1/agent/register [post]
func (DxAgentProfileApi) Register(c *gin.Context) {

	var req request.DxCreateAgentProfileRequest
	var headerImageFile = (*multipart.FileHeader)(nil)
	// Handle both JSON and multipart/form-data
	if c.ContentType() == "multipart/form-data" {
		// Bind form data
		if err := c.ShouldBind(&req); err != nil {
			response.FailWithMessage(err.Error(), c)
			return
		}
		// Try to get the image file (try both common field names)
		if file, err := c.FormFile("headerImage"); err == nil {
			headerImageFile = file
		} else if file, err := c.FormFile("headerImg"); err == nil {
			headerImageFile = file
		} else if file, err := c.FormFile("header_img"); err == nil {
			headerImageFile = file
		}
	} else if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	// Call the service with both request and optional image
	user, err := dxAgentProfileService.Register(req, headerImageFile)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	response.OkWithData(user, c)
}

// Login
// @Tags SystemAuth
// @Summary System login
// @Produce application/json
// @Param data body request.DxLoginAgentRequest true "login"
// @Success 200 {object} response.Response
// @Router /api/v1/system/login [post]
func (dxAgentProfileApi DxAgentProfileApi) Login(c *gin.Context) {
	var req request.DxLoginAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	user, err := dxAgentProfileService.Login(req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	dxAgentProfileApi.TokenNext(c, middleware.ManageTokenHeader, *user)
}

// helper func

// TokenNext 登录以后签发jwt
func (dxAgentProfileApi *DxAgentProfileApi) TokenNext(c *gin.Context, user dxapi.DxAgentProfile) {
	token, claims, err := utils.BxLoginToken(&user)
	if err != nil {
		global.GVA_LOG.Error("获取token失败!", zap.Error(err))
		response.FailWithMessage("获取token失败", c)
		return
	}
	if !global.GVA_CONFIG.System.UseMultipoint {
		utils.SetToken(c, token, int(claims.RegisteredClaims.ExpiresAt.Unix()-time.Now().Unix()))
		response.OkWithDetailed(bxapiRes.LoginResponse{
			User:      user,
			Token:     token,
			ExpiresAt: claims.RegisteredClaims.ExpiresAt.Unix() * 1000,
		}, "登录成功", c)
		return
	}

	if jwtStr, err := bxJwtService.GetRedisJWT(user.Name); err == redis.Nil {
		if err := bxJwtService.SetRedisJWT(token, user.Name); err != nil {
			global.GVA_LOG.Error("设置登录状态失败!", zap.Error(err))
			response.FailWithMessage("设置登录状态失败", c)
			return
		}
		utils.SetToken(c, token, int(claims.RegisteredClaims.ExpiresAt.Unix()-time.Now().Unix()))
		response.OkWithDetailed(bxapiRes.LoginResponse{
			User:      user,
			Token:     token,
			ExpiresAt: claims.RegisteredClaims.ExpiresAt.Unix() * 1000,
		}, "登录成功", c)
	} else if err != nil {
		global.GVA_LOG.Error("设置登录状态失败!", zap.Error(err))
		response.FailWithMessage("设置登录状态失败", c)
	} else {
		var blackJWT system.JwtBlacklist
		blackJWT.Jwt = jwtStr
		if err := bxJwtService.JsonInBlacklist(blackJWT); err != nil {
			response.FailWithMessage("jwt作废失败", c)
			return
		}
		if err := bxJwtService.SetRedisJWT(token, user.GetUsername()); err != nil {
			response.FailWithMessage("设置登录状态失败", c)
			return
		}
		utils.SetToken(c, token, int(claims.RegisteredClaims.ExpiresAt.Unix()-time.Now().Unix()))
		response.OkWithDetailed(bxapiRes.LoginResponse{
			User:      user,
			Token:     token,
			ExpiresAt: claims.RegisteredClaims.ExpiresAt.Unix() * 1000,
		}, "登录成功", c)
	}
}
