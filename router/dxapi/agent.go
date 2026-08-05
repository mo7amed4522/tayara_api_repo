package dxapi

import "github.com/gin-gonic/gin"

type AgentRouter struct{}

// InitAgentRouter registers the agent identity endpoints:
//   - public  : register + login
//   - manage  : own profile + change-password (A-token)
//   - system  : admin list/detail/approve/reject/suspend/note (S-token)
func (r AgentRouter) InitAgentRouter(publicGroup, manageGroup, systemGroup *gin.RouterGroup) {
	public := publicGroup.Group("api/v1/agent")
	manage := manageGroup.Group("api/v1/agent")
	system := systemGroup.Group("api/v1/agent")

	{
		public.POST("register", agentApi.Register)
		public.POST("login", agentApi.Login)
	}
	{
		manage.GET("profile", agentApi.GetProfile)
		manage.POST("change-password", agentApi.ChangePassword)
	}
	{
		system.GET("admin", agentApi.ListAgents)
		system.GET("admin/detail", agentApi.GetAdminDetail)
		system.POST("admin/approve", agentApi.Approve)
		system.POST("admin/reject", agentApi.Reject)
		system.POST("admin/suspend", agentApi.Suspend)
		system.POST("admin/note", agentApi.AddNote)
	}
}
