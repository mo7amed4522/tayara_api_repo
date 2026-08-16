package bxapi

import "github.com/gin-gonic/gin"

type AgentDetailRouter struct{}

// InitAgentDetailRouter registers the customer-facing agent detail endpoint on the public group.
func (r AgentDetailRouter) InitAgentDetailRouter(publicGroup *gin.RouterGroup) {
	customer := publicGroup.Group("api/v1/customer")
	{
		customer.GET("agent/detail", agentDetailApi.CustomerAgentDetail)
	}
}
