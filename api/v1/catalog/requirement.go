package catalog

import (
	"github.com/gin-gonic/gin"
	catReq "github.com/mo7amed4522/tayara_api/model/catalog/request"
	commonReq "github.com/mo7amed4522/tayara_api/model/common/request"
	"github.com/mo7amed4522/tayara_api/model/common/response"
	"github.com/mo7amed4522/tayara_api/utils"
)

type RequirementApi struct{}

// SetProfiles
//
//	@Tags		Catalog
//	@Summary	Agent sets custom-trip requirement profiles for a template+country
//	@Produce	application/json
//	@Param		data	body		catReq.SetRequirementProfilesReq	true	"profiles payload"
//	@Success	200		{object}	response.Response
//	@Router		/api/v1/catalog/templates/requirements [post]
func (RequirementApi) SetProfiles(c *gin.Context) {
	agentID := utils.GetAgentID(c)
	var req catReq.SetRequirementProfilesReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := catalogService.RequirementServices.SetProfiles(agentID, req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("requirement profiles saved", c)
}

// GetProfiles
//
//	@Tags		Catalog
//	@Summary	Agent gets custom-trip requirement profiles for a template+country
//	@Produce	application/json
//	@Param		data	query		catReq.SetRequirementProfilesReq	true	"template + country"
//	@Param		page		query		int	false	"page number"	default(1)
//	@Param		pageSize	query		int	false	"page size"	default(10)
//	@Success	200		{object}	response.Response
//	@Router		/api/v1/catalog/templates/requirements [get]
func (RequirementApi) GetProfiles(c *gin.Context) {
	agentID := utils.GetAgentID(c)
	var req catReq.SetRequirementProfilesReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	var pageInfo commonReq.PageInfo
	if err := c.ShouldBindQuery(&pageInfo); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, totalPage, err := catalogService.RequirementServices.GetProfiles(
		agentID, req.TemplateID, req.CountryID, pageInfo.Page, pageInfo.PageSize,
	)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if pageInfo.Page <= 0 {
		pageInfo.Page = 1
	}
	if pageInfo.PageSize <= 0 {
		pageInfo.PageSize = 10
	}
	response.OkWithDetailed(response.PageResult{
		List: list, Total: total, TotalPage: totalPage,
		Page: pageInfo.Page, PageSize: pageInfo.PageSize,
	}, "ok", c)
}
