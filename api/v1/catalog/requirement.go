package catalog

import (
	"github.com/gin-gonic/gin"
	catReq "github.com/mo7amed4522/tayara_api/model/catalog/request"
	"github.com/mo7amed4522/tayara_api/model/common/response"
	"github.com/mo7amed4522/tayara_api/utils"
)

type RequirementApi struct{}

// SetProfiles
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
//	@Tags		Catalog
//	@Summary	Agent gets custom-trip requirement profiles for a template+country
//	@Produce	application/json
//	@Param		data	query		catReq.SetRequirementProfilesReq	true	"template + country"
//	@Success	200		{object}	response.Response
//	@Router		/api/v1/catalog/templates/requirements [get]
func (RequirementApi) GetProfiles(c *gin.Context) {
	agentID := utils.GetAgentID(c)
	var req catReq.SetRequirementProfilesReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, err := catalogService.RequirementServices.GetProfiles(agentID, req.TemplateID, req.CountryID)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(list, c)
}
