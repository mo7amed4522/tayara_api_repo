package booking

import (
	"github.com/gin-gonic/gin"
	bkReq "github.com/mo7amed4522/tayara_api/model/booking/request"
	commonReq "github.com/mo7amed4522/tayara_api/model/common/request"
	"github.com/mo7amed4522/tayara_api/model/common/response"
	"github.com/mo7amed4522/tayara_api/utils"
)

type DisputeApi struct{}

// Create
//
//	@Tags		Booking
//	@Summary	User or agent opens a dispute on a booking
//	@Produce	application/json
//	@Param		data	body		bkReq.CreateDisputeReq	true	"dispute payload"
//	@Success	200		{object}	response.Response
//	@Router		/api/v1/booking/disputes [post]
func (DisputeApi) Create(c *gin.Context) {
	raisedBy := "user"
	var raiserID uint
	if id := utils.GetBxUserID(c); id > 0 {
		raiserID = id
	} else if id := utils.GetAgentID(c); id > 0 {
		raiserID = id
		raisedBy = "agent"
	}
	if raiserID == 0 {
		response.FailWithMessage("unauthorized", c)
		return
	}
	var req bkReq.CreateDisputeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	dispute, err := bookingService.DisputeServices.Create(raisedBy, raiserID, req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(dispute, c)
}

// List
//
//	@Tags		Booking
//	@Summary	Admin lists disputes (filterable by booking/status)
//	@Produce	application/json
//	@Param		data	query		bkReq.DisputePageReq	true	"filter"
//	@Success	200		{object}	response.Response
//	@Router		/api/v1/booking/admin/disputes [get]
func (DisputeApi) List(c *gin.Context) {
	var req bkReq.DisputePageReq
	_ = c.ShouldBindQuery(&req)
	list, total, totalPage, err := bookingService.DisputeServices.List(req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(response.PageResult{List: list, Total: total, TotalPage: totalPage, Page: req.Page, PageSize: req.PageSize}, "ok", c)
}

// Resolve
//
//	@Tags		Booking
//	@Summary	Admin resolves/closes a dispute
//	@Produce	application/json
//	@Param		data	body		bkReq.ResolveDisputeReq	true	"resolve payload"
//	@Success	200		{object}	response.Response
//	@Router		/api/v1/booking/admin/disputes/resolve [post]
func (DisputeApi) Resolve(c *gin.Context) {
	adminID := utils.GetUserID(c)
	var req bkReq.ResolveDisputeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	dispute, err := bookingService.DisputeServices.Resolve(adminID, req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(dispute, c)
}

// ListForBooking
//
//	@Tags		Booking
//	@Summary	User/agent lists disputes for their own booking
//	@Produce	application/json
//	@Param		id			query		int	true	"booking id"
//	@Param		page		query		int	false	"page number"	default(1)
//	@Param		pageSize	query		int	false	"page size"		default(10)
//	@Success	200			{object}	response.Response
//	@Router		/api/v1/booking/disputes [get]
func (DisputeApi) ListForBooking(c *gin.Context) {
	var idReq commonReq.GetById
	if err := c.ShouldBindQuery(&idReq); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	var pageInfo commonReq.PageInfo
	if err := c.ShouldBindQuery(&pageInfo); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, totalPage, err := bookingService.DisputeServices.ListForBooking(idReq.Uint(), pageInfo.Page, pageInfo.PageSize)
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
	response.OkWithDetailed(response.PageResult{List: list, Total: total, TotalPage: totalPage, Page: pageInfo.Page, PageSize: pageInfo.PageSize}, "ok", c)
}
