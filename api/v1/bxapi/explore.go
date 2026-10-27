package bxapi

import (
	"strconv"

	"github.com/gin-gonic/gin"
	bxapiReq "github.com/mo7amed4522/tayara_api/model/bxapi/request"
	response "github.com/mo7amed4522/tayara_api/model/common/response"
	engageReq "github.com/mo7amed4522/tayara_api/model/engage/request"
	"github.com/mo7amed4522/tayara_api/utils"
)

// ExploreApi handles explore topic endpoints for public/agent APIs
type ExploreApi struct{}

// parseUint converts string to uint
func parseUint(s string) uint {
	id, _ := strconv.ParseUint(s, 10, 32)
	return uint(id)
}

// ListExploreTopics
//
//	@Tags		Customer
//	@Summary	Public explore topics listing
//	@Produce	application/json
//	@Param		page		query		int	false	"page number"	default(1)
//	@Param		pageSize	query		int	false	"page size"		default(20)
//	@Success	200			{object}	response.Response
//	@Router		/api/v1/customer/explore/topics [get]
func (ExploreApi) ListExploreTopics(c *gin.Context) {
	var req bxapiReq.CustomerExploreTopicsReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	list, err := bxapiSvc.ExploreServices.ListExploreTopics(req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(list, c)
}

// GetExploreTopicDetail
//
//	@Tags		Customer
//	@Summary	Public explore topic detail with featured packages
//	@Produce	application/json
//	@Param		code	path		string	true	"topic code"
//	@Success	200			{object}	response.Response
//	@Router		/api/v1/customer/explore/topics/:code [get]
func (ExploreApi) GetExploreTopicDetail(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		response.FailWithMessage("topic code is required", c)
		return
	}

	detail, err := bxapiSvc.ExploreServices.GetExploreTopicDetail(code)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(detail, c)
}

// GetExploreTopicPackages
//
//	@Tags		Customer
//	@Summary	Public packages for an explore topic
//	@Produce	application/json
//	@Param		code		path		string	true	"topic code"
//	@Param		category	query		string	false	"category filter"
//	@Param		page		query		int		false	"page number"	default(1)
//	@Param		pageSize	query		int		false	"page size"		default(12)
//	@Success	200			{object}	response.Response
//	@Router		/api/v1/customer/explore/topics/:code/packages [get]
func (ExploreApi) GetExploreTopicPackages(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		response.FailWithMessage("topic code is required", c)
		return
	}

	var req bxapiReq.CustomerExploreTopicPackagesReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 12
	}

	result, err := bxapiSvc.ExploreServices.GetExploreTopicPackages(code, req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(result, c)
}

// AgentGetMyTopics
//
//	@Tags		Agent
//	@Summary	Agent's topics with their package counts (paginated)
//	@Produce	application/json
//	@Security	BearerAuth
//	@Param		page		query		int	false	"page number"	default(1)
//	@Param		pageSize	query		int	false	"page size"		default(20)
//	@Success	200			{object}	response.Response
//	@Router		/api/v1/agent/explore/topics [get]
func (ExploreApi) AgentGetMyTopics(c *gin.Context) {
	agentID := utils.GetAgentID(c)
	if agentID == 0 {
		response.FailWithMessage("unauthorized", c)
		return
	}

	// Parse pagination parameters
	page := 1
	pageSize := 20
	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if ps := c.Query("pageSize"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 {
			pageSize = parsed
		}
	}

	result, err := bxapiSvc.ExploreServices.AgentGetMyTopics(agentID, page, pageSize)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(result, c)
}

// AgentAddPackageToTopics
//
//	@Tags		Agent
//	@Summary	Add package to explore topics
//	@Produce	application/json
//	@Security	BearerAuth
//	@Param		id		path		int		true	"package ID"
//	@Param		body	body		engageReq.AgentPackageTopicsReq	true	"topic IDs"
//	@Success	200		{object}	response.Response
//	@Router		/api/v1/agent/explore/packages/:id/topics [post]
func (ExploreApi) AgentAddPackageToTopics(c *gin.Context) {
	agentID := utils.GetAgentID(c)
	if agentID == 0 {
		response.FailWithMessage("unauthorized", c)
		return
	}

	packageIDStr := c.Param("id")
	packageID := parseUint(packageIDStr)

	var req engageReq.AgentPackageTopicsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	err := bxapiSvc.ExploreServices.AgentAddPackageToTopics(agentID, packageID, req.TopicIDs)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("package added to topics", c)
}

// AgentRemovePackageFromTopic
//
//	@Tags		Agent
//	@Summary	Remove package from an explore topic
//	@Produce	application/json
//	@Security	BearerAuth
//	@Param		id		path		int	true	"package ID"
//	@Param		topicId	path		int	true	"topic ID"
//	@Success	200		{object}	response.Response
//	@Router		/api/v1/agent/explore/packages/:id/topics/:topicId [delete]
func (ExploreApi) AgentRemovePackageFromTopic(c *gin.Context) {
	agentID := utils.GetAgentID(c)
	if agentID == 0 {
		response.FailWithMessage("unauthorized", c)
		return
	}

	packageIDStr := c.Param("id")
	packageID := parseUint(packageIDStr)

	topicIDStr := c.Param("topicId")
	topicID := parseUint(topicIDStr)

	err := bxapiSvc.ExploreServices.AgentRemovePackageFromTopic(agentID, packageID, topicID)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("package removed from topic", c)
}

// AgentGetPackageTopics
//
//	@Tags		Agent
//	@Summary	Get explore topics for a package
//	@Produce	application/json
//	@Security	BearerAuth
//	@Param		id	path		int	true	"package ID"
//	@Success	200		{object}	response.Response
//	@Router		/api/v1/agent/explore/packages/:id/topics [get]
func (ExploreApi) AgentGetPackageTopics(c *gin.Context) {
	agentID := utils.GetAgentID(c)
	if agentID == 0 {
		response.FailWithMessage("unauthorized", c)
		return
	}

	packageIDStr := c.Param("id")
	packageID := parseUint(packageIDStr)

	topicIDs, err := bxapiSvc.ExploreServices.AgentGetPackageTopics(agentID, packageID)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(topicIDs, c)
}
