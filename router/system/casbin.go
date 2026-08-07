package system

import "github.com/gin-gonic/gin"

type CasbinRouter struct{}

// InitCasbinRouter registers the RBAC permission-management endpoints (admin-only):
//   - GET  /api/v1/system/casbin/policies  list a role's route-authority matrix
//   - PUT  /api/v1/system/casbin/policies  replace a role's full matrix (sync)
//   - PUT  /api/v1/system/casbin/policy    update a single route rule
func (r CasbinRouter) InitCasbinRouter(privateGroup *gin.RouterGroup) {
	group := privateGroup.Group("api/v1/system/casbin")
	{
		group.GET("policies", casbinApi.ListPolicies)
		group.PUT("policies", casbinApi.SyncPolicies)
		group.PUT("policy", casbinApi.UpdatePolicy)
	}
}
