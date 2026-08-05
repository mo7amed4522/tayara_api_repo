package payment

import (
	"github.com/gin-gonic/gin"
	"github.com/mo7amed4522/tayara_api/model/common/response"
	paymentReq "github.com/mo7amed4522/tayara_api/model/payment/request"
	"github.com/mo7amed4522/tayara_api/utils"
)

type CommissionApi struct{}

// Override
// @Tags Payment
// @Summary Admin overrides commission % on a captured booking (reason required, audited)
// @Produce application/json
// @Param data body paymentReq.CommissionOverrideReq true "override payload"
// @Success 200 {object} response.Response
// @Router /api/v1/payment/commission/override [post]
func (CommissionApi) Override(c *gin.Context) {
	adminID := utils.GetUserID(c)
	var req paymentReq.CommissionOverrideReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	rec, err := paymentService.PaymentServices.OverrideCommission(adminID, req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(rec, c)
}

// List
// @Tags Payment
// @Summary Admin lists commission records (filterable)
// @Produce application/json
// @Param data query paymentReq.CommissionPageReq true "filter"
// @Success 200 {object} response.Response
// @Router /api/v1/payment/commission [get]
func (CommissionApi) List(c *gin.Context) {
	var req paymentReq.CommissionPageReq
	_ = c.ShouldBindQuery(&req)
	list, total, err := paymentService.PaymentServices.ListCommission(0, req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(response.PageResult{List: list, Total: total, Page: req.Page, PageSize: req.PageSize}, "ok", c)
}

// AgentList
// @Tags Payment
// @Summary Agent lists their own commission records
// @Produce application/json
// @Param data query paymentReq.CommissionPageReq true "filter"
// @Success 200 {object} response.Response
// @Router /api/v1/payment/commission/manage [get]
func (CommissionApi) AgentList(c *gin.Context) {
	agentID := utils.GetAgentID(c)
	var req paymentReq.CommissionPageReq
	_ = c.ShouldBindQuery(&req)
	list, total, err := paymentService.PaymentServices.ListCommission(agentID, req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(response.PageResult{List: list, Total: total, Page: req.Page, PageSize: req.PageSize}, "ok", c)
}
