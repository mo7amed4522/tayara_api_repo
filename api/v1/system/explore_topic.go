package system

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mo7amed4522/tayara_api/model/engage/request"
	commonResp "github.com/mo7amed4522/tayara_api/model/common/response"
)

// ExploreTopicApi handles admin endpoints for explore topics
type ExploreTopicApi struct{}

var exploreTopicApi = ExploreTopicApi{}

// AdminListTopics
//	@Tags		System
//	@Summary	Admin list explore topics
//	@Produce	application/json
//	@Security	BearerAuth
//	@Param		page		query		int		false	"page number"	default(1)
//	@Param		pageSize	query		int		false	"page size"		default(20)
//	@Param		isActive	query		bool	false	"filter by active status"
//	@Param		keyword		query		string	false	"search keyword"
//	@Success	200			{object}	commonResp.Response
//	@Router		/api/v1/admin/explore/topics [get]
func (ExploreTopicApi) AdminListTopics(c *gin.Context) {
	var req request.AdminExploreTopicPageReq
	if err := c.ShouldBindQuery(&req); err != nil {
		commonResp.FailWithMessage(err.Error(), c)
		return
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	result, err := systemSvc.ExploreTopicServices.AdminListTopics(req)
	if err != nil {
		commonResp.FailWithMessage(err.Error(), c)
		return
	}
	commonResp.OkWithData(result, c)
}

// AdminGetTopic
//	@Tags		System
//	@Summary	Admin get explore topic by ID
//	@Produce	application/json
//	@Security	BearerAuth
//	@Param		id	path		int	true	"topic ID"
//	@Success	200		{object}	commonResp.Response
//	@Router		/api/v1/admin/explore/topics/:id [get]
func (ExploreTopicApi) AdminGetTopic(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		commonResp.FailWithMessage("invalid topic ID", c)
		return
	}

	topic, err := systemSvc.ExploreTopicServices.AdminGetTopic(uint(id))
	if err != nil {
		commonResp.FailWithMessage(err.Error(), c)
		return
	}
	commonResp.OkWithData(topic, c)
}

// AdminCreateTopic
//	@Tags		System
//	@Summary	Admin create explore topic
//	@Produce	application/json
//	@Security	BearerAuth
//	@Param		body	body		request.ExploreTopicReq	true	"topic data"
//	@Success	200		{object}	commonResp.Response
//	@Router		/api/v1/admin/explore/topics [post]
func (ExploreTopicApi) AdminCreateTopic(c *gin.Context) {
	var req request.ExploreTopicReq
	if err := c.ShouldBindJSON(&req); err != nil {
		commonResp.FailWithMessage(err.Error(), c)
		return
	}

	topic, err := systemSvc.ExploreTopicServices.AdminCreateTopic(req)
	if err != nil {
		commonResp.FailWithMessage(err.Error(), c)
		return
	}
	commonResp.OkWithData(topic, c)
}

// AdminUpdateTopic
//	@Tags		System
//	@Summary	Admin update explore topic
//	@Produce	application/json
//	@Security	BearerAuth
//	@Param		id	path		int	true	"topic ID"
//	@Param		body	body		request.ExploreTopicReq	true	"topic data"
//	@Success	200		{object}	commonResp.Response
//	@Router		/api/v1/admin/explore/topics/:id [put]
func (ExploreTopicApi) AdminUpdateTopic(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		commonResp.FailWithMessage("invalid topic ID", c)
		return
	}

	var req request.ExploreTopicReq
	if err := c.ShouldBindJSON(&req); err != nil {
		commonResp.FailWithMessage(err.Error(), c)
		return
	}

	topic, err := systemSvc.ExploreTopicServices.AdminUpdateTopic(uint(id), req)
	if err != nil {
		commonResp.FailWithMessage(err.Error(), c)
		return
	}
	commonResp.OkWithData(topic, c)
}

// AdminDeleteTopic
//	@Tags		System
//	@Summary	Admin delete explore topic
//	@Produce	application/json
//	@Security	BearerAuth
//	@Param		id	path		int	true	"topic ID"
//	@Success	200		{object}	commonResp.Response
//	@Router		/api/v1/admin/explore/topics/:id [delete]
func (ExploreTopicApi) AdminDeleteTopic(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		commonResp.FailWithMessage("invalid topic ID", c)
		return
	}

	err = systemSvc.ExploreTopicServices.AdminDeleteTopic(uint(id))
	if err != nil {
		commonResp.FailWithMessage(err.Error(), c)
		return
	}
	commonResp.OkWithMessage("topic deleted", c)
}

// AdminReorderTopics
//	@Tags		System
//	@Summary	Admin reorder explore topics
//	@Produce	application/json
//	@Security	BearerAuth
//	@Param		body	body		[]request.TopicOrderReq	true	"topic order list"
//	@Success	200		{object}	commonResp.Response
//	@Router		/api/v1/admin/explore/topics/reorder [put]
func (ExploreTopicApi) AdminReorderTopics(c *gin.Context) {
	var req []request.TopicOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		commonResp.FailWithMessage(err.Error(), c)
		return
	}

	err := systemSvc.ExploreTopicServices.AdminReorderTopics(req)
	if err != nil {
		commonResp.FailWithMessage(err.Error(), c)
		return
	}
	commonResp.OkWithMessage("topics reordered", c)
}

// AdminManageTopicPackages
//	@Tags		System
//	@Summary	Admin manage packages in explore topic
//	@Produce	application/json
//	@Security	BearerAuth
//	@Param		id		path		int	true	"topic ID"
//	@Param		body	body		request.AdminManageTopicPackagesReq	true	"package IDs and replace flag"
//	@Success	200		{object}	commonResp.Response
//	@Router		/api/v1/admin/explore/topics/:id/packages [post]
func (ExploreTopicApi) AdminManageTopicPackages(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		commonResp.FailWithMessage("invalid topic ID", c)
		return
	}

	var req request.AdminManageTopicPackagesReq
	if err := c.ShouldBindJSON(&req); err != nil {
		commonResp.FailWithMessage(err.Error(), c)
		return
	}

	err = systemSvc.ExploreTopicServices.AdminManageTopicPackages(uint(id), req)
	if err != nil {
		commonResp.FailWithMessage(err.Error(), c)
		return
	}
	commonResp.OkWithMessage("topic packages updated", c)
}

// AdminGetTopicPackages
//	@Tags		System
//	@Summary	Admin get packages in explore topic
//	@Produce	application/json
//	@Security	BearerAuth
//	@Param		id		path		int	true	"topic ID"
//	@Param		page	query		int	false	"page number"	default(1)
//	@Param		pageSize	query		int	false	"page size"		default(20)
//	@Success	200		{object}	commonResp.Response
//	@Router		/api/v1/admin/explore/topics/:id/packages [get]
func (ExploreTopicApi) AdminGetTopicPackages(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		commonResp.FailWithMessage("invalid topic ID", c)
		return
	}

	pageStr := c.DefaultQuery("page", "1")
	page, _ := strconv.Atoi(pageStr)
	if page <= 0 {
		page = 1
	}

	pageSizeStr := c.DefaultQuery("pageSize", "20")
	pageSize, _ := strconv.Atoi(pageSizeStr)
	if pageSize <= 0 {
		pageSize = 20
	}

	result, err := systemSvc.ExploreTopicServices.AdminGetTopicPackages(uint(id), page, pageSize)
	if err != nil {
		commonResp.FailWithMessage(err.Error(), c)
		return
	}
	commonResp.OkWithData(result, c)
}