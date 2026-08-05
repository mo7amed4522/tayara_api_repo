package analytics

import (
	"github.com/gin-gonic/gin"
	"github.com/mo7amed4522/tayara_api/model/common/response"
	"github.com/mo7amed4522/tayara_api/utils"
)

type AnalyticsApi struct{}

// AdminUsers
// @Tags Analytics
// @Summary Admin Users dashboard card
// @Produce application/json
// @Success 200 {object} response.Response
// @Router /api/v1/analytics/admin/users [get]
func (AnalyticsApi) AdminUsers(c *gin.Context) {
	card, err := analyticsService.AnalyticsServices.AdminUsersCard()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(card, c)
}

// AdminAgents
// @Tags Analytics
// @Summary Admin Agents dashboard card
// @Produce application/json
// @Success 200 {object} response.Response
// @Router /api/v1/analytics/admin/agents [get]
func (AnalyticsApi) AdminAgents(c *gin.Context) {
	card, err := analyticsService.AnalyticsServices.AdminAgentsCard()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(card, c)
}

// AdminBusiness
// @Tags Analytics
// @Summary Admin Business dashboard card (GMV, commission, payouts, refunds, net profit)
// @Produce application/json
// @Success 200 {object} response.Response
// @Router /api/v1/analytics/admin/business [get]
func (AnalyticsApi) AdminBusiness(c *gin.Context) {
	card, err := analyticsService.AnalyticsServices.AdminBusinessCard()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(card, c)
}

// AgentOverview
// @Tags Analytics
// @Summary Agent dashboard card (revenue, bookings, views, avg confirmation time)
// @Produce application/json
// @Success 200 {object} response.Response
// @Router /api/v1/analytics/agent [get]
func (AnalyticsApi) AgentOverview(c *gin.Context) {
	agentID := utils.GetAgentID(c)
	card, err := analyticsService.AnalyticsServices.AgentCard(agentID)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(card, c)
}
