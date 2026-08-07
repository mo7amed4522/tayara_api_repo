package system

import (
	"github.com/gin-gonic/gin"
)

type SysAdminProfileRouter struct{}

func (sysAdminProfileRouter SysAdminProfileRouter) InitSysAdminProfileRouter(publicGroup, privateGroup *gin.RouterGroup) {

	public := publicGroup.Group("api/v1/system")
	private := privateGroup.Group("api/v1/system")

	{
		public.POST("register", sysAdminProfileApi.Register)
		public.POST("login", sysAdminProfileApi.Login)
		public.POST("forgot-password", sysAdminProfileApi.ForgotPassword)
		public.POST("reset-password", sysAdminProfileApi.ResetPassword)
	}
	{
		private.POST("changePassword", sysAdminProfileApi.ChangePassword)
		private.GET("jobs", sysAdminProfileApi.JobRegistry)
	}
}
