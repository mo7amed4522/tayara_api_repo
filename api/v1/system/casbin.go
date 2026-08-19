package system

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/model/common/response"
	sysReq "github.com/mo7amed4522/tayara_api/model/system/request"
	"github.com/mo7amed4522/tayara_api/utils"
)

// CasbinApi exposes the RBAC permission-management endpoints (admin). It builds on the existing
// CasbinService (policy persistence via the gorm casbin_rule table) so an admin can view, add and
// replace the route-authority matrix per role — the management half of the Phase-9 enforcement.
type CasbinApi struct{}

// casbinSvc reaches the CasbinService via the system service group (promoted method set).
var casbinSvc = sysAdminProfile.CasbinService

// ListPolicies
//	@Tags		SystemCasbin
//	@Summary	Admin lists the permission matrix (routes) for a role
//	@Produce	application/json
//	@Param		id	query		uint	true	"authority id"
//	@Success	200	{object}	response.Response
//	@Router		/api/v1/system/casbin/policies [get]
func (CasbinApi) ListPolicies(c *gin.Context) {
	// Optional ?id=<authorityId> override (admin acting for another role). If omitted, fall back
	// to the caller's own authority id from the JWT — so a bare authenticated call (just an
	// S-token) returns that role's policies without requiring an explicit id.
	var authID uint
	idParam := c.Query("id")
	if idParam != "" {
		parsed, err := strconv.ParseUint(idParam, 10, 64)
		if err != nil {
			response.FailWithMessage("invalid authority id", c)
			return
		}
		authID = uint(parsed)
	} else {
		authID = utils.GetUserAuthorityId(c)
	}
	policies := casbinSvc.GetPolicyPathByAuthorityId(authID)
	response.OkWithData(gin.H{
		"authorityId": authID,
		"policies":    policies,
	}, c)
}

// SyncPolicies
//	@Tags		SystemCasbin
//	@Summary	Admin replaces the route-authority matrix for a role
//	@Produce	application/json
//	@Param		data	body		sysReq.CasbinInReceive	true	"authority + policies"
//	@Success	200		{object}	response.Response
//	@Router		/api/v1/system/casbin/policies [put]
func (CasbinApi) SyncPolicies(c *gin.Context) {
	var req sysReq.CasbinSyncReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	rules := [][]string{}
	for _, p := range req.Policies {
		rules = append(rules, []string{
			strconv.FormatUint(uint64(req.AuthorityID), 10),
			p.Path,
			p.Method,
		})
	}
	if err := casbinSvc.SyncPolicy(global.GVA_DB, strconv.FormatUint(uint64(req.AuthorityID), 10), rules); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	_ = casbinSvc.FreshCasbin()
	response.OkWithDetailed(gin.H{"authorityId": req.AuthorityID, "count": len(rules)}, "policies synchronized", c)
}

// UpdatePolicy
//	@Tags		SystemCasbin
//	@Summary	Admin updates a single route rule for a role
//	@Produce	application/json
//	@Param		data	body		sysReq.CasbinUpdateReq	true	"old + new path/method"
//	@Success	200		{object}	response.Response
//	@Router		/api/v1/system/casbin/policy [put]
func (CasbinApi) UpdatePolicy(c *gin.Context) {
	var req sysReq.CasbinUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := casbinSvc.UpdateCasbinApi(req.OldPath, req.NewPath, req.OldMethod, req.NewMethod); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("policy updated", c)
}
