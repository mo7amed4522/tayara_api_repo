package bxapi

import (
	"github.com/gin-gonic/gin"
	bxapiReq "github.com/mo7amed4522/tayara_api/model/bxapi/request"
	"github.com/mo7amed4522/tayara_api/model/common/response"
)

type AgentDetailApi struct{}

// CustomerAgentDetail
// @Tags Customer
// @Summary Customer agent detail page (public)
// @Produce application/json
// @Param agentId query int true "agent id"
// @Param destinationCountryId query int false "filter packages by destination country"
// @Param page query int false "page number" default(1)
// @Param pageSize query int false "page size" default(12)
// @Success 200 {object} response.Response
// @Router /api/v1/customer/agent/detail [get]
func (api AgentDetailApi) CustomerAgentDetail(c *gin.Context) {
	var req bxapiReq.CustomerAgentDetailReq
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
	if req.AgentID == 0 {
		response.FailWithMessage("agentId is required", c)
		return
	}
	result, err := agentDetailService.AgentDetailServices.CustomerAgentDetail(req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(result, c)
}
