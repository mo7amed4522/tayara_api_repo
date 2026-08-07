package bxapi

import "github.com/gin-gonic/gin"

type UserRouter struct{}

// InitUserRouter registers the user identity endpoints:
//   - public : register + login
//   - bx     : own profile + change-password (X-token)
//   - system : admin list + block/unblock (S-token)
func (r UserRouter) InitUserRouter(publicGroup, bxGroup, systemGroup *gin.RouterGroup) {
	public := publicGroup.Group("api/v1/user")
	bx := bxGroup.Group("api/v1/user")
	system := systemGroup.Group("api/v1/user")

	{
		public.POST("register", userApi.Register)
		public.POST("login", userApi.Login)
		public.POST("firebase-login", userApi.FirebaseLogin)
	}
	{
		bx.GET("profile", userApi.GetProfile)
		bx.POST("change-password", userApi.ChangePassword)
	}
	{
		system.GET("admin", userApi.ListUsers)
		system.POST("admin/block", userApi.Block)
	}
}
