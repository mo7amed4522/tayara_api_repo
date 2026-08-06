package system

import (
	"mime/multipart"

	"github.com/gin-gonic/gin"
	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/middleware"
	"github.com/mo7amed4522/tayara_api/model/common/response"
	"github.com/mo7amed4522/tayara_api/model/system"
	"github.com/mo7amed4522/tayara_api/model/system/request"
	req "github.com/mo7amed4522/tayara_api/model/system/request"
	res "github.com/mo7amed4522/tayara_api/model/system/response"
	"github.com/mo7amed4522/tayara_api/utils"
)

type SysAdminProfileApi struct{}

// Register
// @Tags SystemAuth
// @Summary Register system admin (requires X-Admin-Key header)
// @Produce application/json
// @Param X-Admin-Key header string true "Super-admin secret key"
// @Param data body req.SysRegisterReq true "register"
// @Success 200 {object} response.Response
// @Router /api/v1/system/register [post]
func (SysAdminProfileApi) Register(c *gin.Context) {
	// Check admin key if required
	superAdminKey := global.GVA_CONFIG.System.SuperAdminKey
	if superAdminKey != "" {
		provided := c.GetHeader("X-Admin-Key")
		if provided != superAdminKey {
			response.FailWithMessage("Unauthorized: invalid admin key", c)
			return
		}
	}
	var req request.SysRegisterReq
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
	user, err := sysAdminProfile.Register(req, headerImageFile)
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
// @Param data body req.SysLoginReq true "login"
// @Success 200 {object} response.Response
// @Router /api/v1/system/login [post]
func (sysAdminProfileApi SysAdminProfileApi) Login(c *gin.Context) {
	var req req.SysLoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	user, err := sysAdminProfile.Login(req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	sysAdminProfileApi.TokenNext(c, middleware.SystemTokenHeader, *user)
}

// ChangePassword
// @Tags SystemAuth
// @Summary Change system password
// @Security ApiKeyAuth
// @Produce application/json
// @Param data body req.SysChangePasswordReq true "password"
// @Success 200 {object} response.Response
// @Router /api/v1/system/change-password [post]
func (sysAdminProfileApi SysAdminProfileApi) ChangePassword(c *gin.Context) {
	var req req.SysChangePasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	req.ID = utils.GetUserID(c)
	if err := sysAdminProfile.ChangePassword(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("Password changed", c)
}

// helper func
func (SysAdminProfileApi) TokenNext(c *gin.Context, headerName string, user system.SysAdminProfile) {
	result, err := middleware.IssueDomainToken(user, headerName)
	if err != nil {
		response.FailWithMessage("Failed to get token", c)
		return
	}
	c.Header(headerName, result.Token)
	middleware.SetDomainTokenCookie(c.Request.Host, headerName, result.Token, middleware.DomainTokenMaxAge(result.ExpiresAt), c.SetCookie)
	response.OkWithDetailed(res.SysLoginResponse{
		Admin:     user,
		Token:     result.Token,
		ExpiresAt: result.ExpiresAt,
	}, "Login success", c)
}
