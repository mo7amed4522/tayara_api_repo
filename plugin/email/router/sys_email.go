package router

import (
	"github.com/gin-gonic/gin"
	"github.com/mo7amed4522/tayara_api/middleware"
	"github.com/mo7amed4522/tayara_api/plugin/email/api"
)

type EmailRouter struct{}

func (s *EmailRouter) InitEmailRouter(Router *gin.RouterGroup) {
	emailRouter := Router.Use(middleware.OperationRecord())
	EmailApi := api.ApiGroupApp.EmailApi.EmailTest
	SendEmail := api.ApiGroupApp.EmailApi.SendEmail
	{
		emailRouter.POST("emailTest", EmailApi)  // Send test email
		emailRouter.POST("sendEmail", SendEmail) // Send email
	}
}
