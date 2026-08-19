package api

import (
	"github.com/gin-gonic/gin"
	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/model/common/response"
	email_response "github.com/mo7amed4522/tayara_api/plugin/email/model/response"
	"github.com/mo7amed4522/tayara_api/plugin/email/service"
	"go.uber.org/zap"
)

type EmailApi struct{}

// EmailTest
//	@Tags		System
//	@Summary	Send test email
//	@Security	ApiKeyAuth
//	@Produce	application/json
//	@Success	200	{string}	string	"{"success":true,"data":{},"msg":"Send success"}"
//	@Router		/email/emailTest [post]
func (s *EmailApi) EmailTest(c *gin.Context) {
	err := service.ServiceGroupApp.EmailTest()
	if err != nil {
		global.GVA_LOG.Error("Send failed!", zap.Error(err))
		response.FailWithMessage("Send failed", c)
		return
	}
	response.OkWithMessage("Send success", c)
}

// SendEmail
//	@Tags		System
//	@Summary	Send email
//	@Security	ApiKeyAuth
//	@Produce	application/json
//	@Param		data	body		email_response.Email	true									"The parameters required for sending email"
//	@Success	200		{string}	string					"{"success":true,"data":{},"msg":"Send success"}"
//	@Router		/email/sendEmail [post]
func (s *EmailApi) SendEmail(c *gin.Context) {
	var email email_response.Email
	err := c.ShouldBindJSON(&email)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	err = service.ServiceGroupApp.SendEmail(email.To, email.Subject, email.Body)
	if err != nil {
		global.GVA_LOG.Error("Send failed!", zap.Error(err))
		response.FailWithMessage("Send failed", c)
		return
	}
	response.OkWithMessage("Send success", c)
}
