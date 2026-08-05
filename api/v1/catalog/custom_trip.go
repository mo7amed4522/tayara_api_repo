package catalog

import (
	"github.com/gin-gonic/gin"
	catalogReq "github.com/mo7amed4522/tayara_api/model/catalog/request"
	commonReq "github.com/mo7amed4522/tayara_api/model/common/request"
	"github.com/mo7amed4522/tayara_api/model/common/response"
	"github.com/mo7amed4522/tayara_api/utils"
)

type CustomTripApi struct{}

// CreateTemplate
// @Tags Catalog
// @Summary Agent creates a custom-trip builder template
// @Produce application/json
// @Param data body catalogReq.CustomTripTemplateReq true "template payload"
// @Success 200 {object} response.Response
// @Router /api/v1/catalog/templates [post]
func (CustomTripApi) CreateTemplate(c *gin.Context) {
	agentID := utils.GetAgentID(c)
	if agentID == 0 {
		response.FailWithMessage("unauthorized agent", c)
		return
	}
	var req catalogReq.CustomTripTemplateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	tpl, err := catalogService.CustomTripServices.CreateTemplate(agentID, req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(tpl, c)
}

// UpdateTemplate
// @Tags Catalog
// @Summary Agent updates one of their own templates
// @Produce application/json
// @Param data body catalogReq.CustomTripTemplateReq true "template payload"
// @Success 200 {object} response.Response
// @Router /api/v1/catalog/templates [put]
func (CustomTripApi) UpdateTemplate(c *gin.Context) {
	agentID := utils.GetAgentID(c)
	var req catalogReq.CustomTripTemplateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	tpl, err := catalogService.CustomTripServices.UpdateTemplate(agentID, req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(tpl, c)
}

// ListOwnTemplates
// @Tags Catalog
// @Summary Agent lists their own custom-trip templates
// @Produce application/json
// @Success 200 {object} response.Response
// @Router /api/v1/catalog/templates [get]
func (CustomTripApi) ListOwnTemplates(c *gin.Context) {
	agentID := utils.GetAgentID(c)
	list, err := catalogService.CustomTripServices.ListOwnTemplates(agentID)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(list, c)
}

// GetOwnTemplate
// @Tags Catalog
// @Summary Agent gets one of their own templates
// @Produce application/json
// @Param id query int true "template id"
// @Success 200 {object} response.Response
// @Router /api/v1/catalog/templates/detail [get]
func (CustomTripApi) GetOwnTemplate(c *gin.Context) {
	agentID := utils.GetAgentID(c)
	var idReq commonReq.GetById
	if err := c.ShouldBindQuery(&idReq); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	tpl, err := catalogService.CustomTripServices.GetOwnTemplate(agentID, uint(idReq.ID))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(tpl, c)
}

// DeleteTemplate
// @Tags Catalog
// @Summary Agent deletes one of their own templates
// @Produce application/json
// @Param id query int true "template id"
// @Success 200 {object} response.Response
// @Router /api/v1/catalog/templates [delete]
func (CustomTripApi) DeleteTemplate(c *gin.Context) {
	agentID := utils.GetAgentID(c)
	var idReq commonReq.GetById
	if err := c.ShouldBindQuery(&idReq); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := catalogService.CustomTripServices.DeleteTemplate(agentID, uint(idReq.ID)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("template deleted", c)
}
