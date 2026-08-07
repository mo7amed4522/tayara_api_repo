package payment

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/mo7amed4522/tayara_api/model/common/response"
	paymentReq "github.com/mo7amed4522/tayara_api/model/payment/request"
	"github.com/mo7amed4522/tayara_api/utils"
)

type PayoutApi struct{}

// PreparePayouts
// @Tags Payment
// @Summary Admin scans bookings and prepares payout records (eligible -> pending)
// @Produce application/json
// @Success 200 {object} response.Response
// @Router /api/v1/payment/payouts/prepare [post]
func (PayoutApi) PreparePayouts(c *gin.Context) {
	_, err := paymentService.PaymentServices.PreparePayouts()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("payouts prepared", c)
}

// PayAgent
// @Tags Payment
// @Summary Admin pays a pending payout for an eligible booking
// @Produce application/json
// @Param data body paymentReq.PayAgentReq true "payout payload"
// @Success 200 {object} response.Response
// @Router /api/v1/payment/payouts/pay [post]
func (PayoutApi) PayAgent(c *gin.Context) {
	adminID := utils.GetUserID(c)
	var req paymentReq.PayAgentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	detail, err := paymentService.PaymentServices.PayAgent(context.Background(), adminID, req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(detail, c)
}

// ListAdminPayouts
// @Tags Payment
// @Summary Admin lists all payouts (paginated)
// @Produce application/json
// @Param data query paymentReq.PayoutPageReq true "page filter"
// @Success 200 {object} response.Response
// @Router /api/v1/payment/payouts [get]
func (PayoutApi) ListAdminPayouts(c *gin.Context) {
	var req paymentReq.PayoutPageReq
	_ = c.ShouldBindQuery(&req)
	list, total, err := paymentService.PaymentServices.ListPayouts(0, req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(response.PageResult{List: list, Total: total, Page: req.Page, PageSize: req.PageSize}, "ok", c)
}

// ListAgentPayouts
// @Tags Payment
// @Summary Agent lists their own payouts (paginated)
// @Produce application/json
// @Param data query paymentReq.PayoutPageReq true "page filter"
// @Success 200 {object} response.Response
// @Router /api/v1/payment/payouts/manage [get]
func (PayoutApi) ListAgentPayouts(c *gin.Context) {
	agentID := utils.GetAgentID(c)
	var req paymentReq.PayoutPageReq
	_ = c.ShouldBindQuery(&req)
	list, total, err := paymentService.PaymentServices.ListPayouts(agentID, req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(response.PageResult{List: list, Total: total, Page: req.Page, PageSize: req.PageSize}, "ok", c)
}

// AgentEarnings
// @Tags Payment
// @Summary Agent views their earnings dashboard
// @Produce application/json
// @Success 200 {object} response.Response
// @Router /api/v1/payment/payouts/earnings [get]
func (PayoutApi) AgentEarnings(c *gin.Context) {
	agentID := utils.GetAgentID(c)
	out, err := paymentService.PaymentServices.AgentEarnings(agentID)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(out, c)
}
