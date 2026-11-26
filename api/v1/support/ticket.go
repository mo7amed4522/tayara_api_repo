package support

import (
	"strconv"

	"github.com/gin-gonic/gin"
	commonReq "github.com/mo7amed4522/tayara_api/model/common/request"
	"github.com/mo7amed4522/tayara_api/model/common/response"
	supportReq "github.com/mo7amed4522/tayara_api/model/support/request"
	supportResp "github.com/mo7amed4522/tayara_api/model/support/response"
)

// TicketApi exposes ticket creation and reads.
type TicketApi struct{}

// Open
//
//	@Tags		Support
//	@Summary	Customer opens a support ticket (creates conversation + welcome message)
//	@Produce	application/json
//	@Param		data	body		supportReq.OpenTicketReq	true	"ticket payload"
//	@Success	200		{object}	response.Response
//	@Router		/api/v1/support/tickets [post]
func (TicketApi) Open(c *gin.Context) {
	var req supportReq.OpenTicketReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithError(response.ValidationErr(err.Error()), c)
		return
	}
	view, err := supportService.TicketServices.Open(currentActor(c), req)
	if err != nil {
		respondErr(err, c)
		return
	}
	response.OkWithData(view, c)
}

// ListOwn
//
//	@Tags		Support
//	@Summary	Customer lists their own support tickets
//	@Produce	application/json
//	@Param		data	query		supportReq.TicketPageReq	true	"filter"
//	@Success	200		{object}	response.Response
//	@Router		/api/v1/support/tickets [get]
func (TicketApi) ListOwn(c *gin.Context) {
	var req supportReq.TicketPageReq
	_ = c.ShouldBindQuery(&req)
	list, total, totalPage, err := supportService.TicketServices.ListOwn(currentActor(c), req)
	if err != nil {
		respondErr(err, c)
		return
	}
	response.OkWithDetailed(response.PageResult{List: list, Total: total, TotalPage: totalPage, Page: req.Page, PageSize: req.PageSize}, "ok", c)
}

// Detail
//
//	@Tags		Support
//	@Summary	Read a ticket the caller may access
//	@Produce	application/json
//	@Param		id	path		int	true	"ticket id"
//	@Success	200	{object}	response.Response
//	@Router		/api/v1/support/tickets/{id} [get]
func (TicketApi) Detail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithError(response.ValidationErr("invalid ticket id"), c)
		return
	}
	view, err := supportService.TicketServices.Detail(currentActor(c), uint(id))
	if err != nil {
		respondErr(err, c)
		return
	}
	response.OkWithData(view, c)
}

// ensure supportResp is referenced (compile-time doc anchor for swagger response models).
var _ = supportResp.TicketView{}

// AdminDetail
//
//	@Tags		Support
//	@Summary	Staff reads a ticket
//	@Produce	application/json
//	@Param		id	path		int	true	"ticket id"
//	@Success	200	{object}	response.Response
//	@Router		/api/v1/support/admin/tickets/{id} [get]
func (TicketApi) AdminDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithError(response.ValidationErr("invalid ticket id"), c)
		return
	}
	view, err := supportService.TicketServices.Detail(currentActor(c), uint(id))
	if err != nil {
		respondErr(err, c)
		return
	}
	response.OkWithData(view, c)
}

// AdminList
//
//	@Tags		Support
//	@Summary	Staff inbox (support: assigned + unassigned, quality: all, admin: all)
//	@Produce	application/json
//	@Param		data	query		supportReq.TicketPageReq	true	"filter"
//	@Success	200		{object}	response.Response
//	@Router		/api/v1/support/admin/tickets [get]
func (TicketApi) AdminList(c *gin.Context) {
	var req supportReq.TicketPageReq
	_ = c.ShouldBindQuery(&req)
	list, total, totalPage, err := supportService.TicketServices.ListAdmin(currentActor(c), req)
	if err != nil {
		respondErr(err, c)
		return
	}
	response.OkWithDetailed(response.PageResult{List: list, Total: total, TotalPage: totalPage, Page: req.Page, PageSize: req.PageSize}, "ok", c)
}

// ConversationDetail
//
//	@Tags		Support
//	@Summary	Read a conversation the caller may access
//	@Produce	application/json
//	@Param		id	path		int	true	"conversation id"
//	@Success	200	{object}	response.Response
//	@Router		/api/v1/support/conversations/{id} [get]
func (TicketApi) ConversationDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithError(response.ValidationErr("invalid conversation id"), c)
		return
	}
	access, ok, err := supportService.AuthorizationServices.CanRead(currentActor(c), uint(id))
	if err != nil {
		respondErr(err, c)
		return
	}
	if !ok {
		response.FailWithError(response.ForbiddenErr("not allowed to read this conversation"), c)
		return
	}
	response.OkWithData(access.Conversation, c)
}

var _ = commonReq.PageInfo{}
