package support

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mo7amed4522/tayara_api/constants"
	"github.com/mo7amed4522/tayara_api/model/common/response"
	supportReq "github.com/mo7amed4522/tayara_api/model/support/request"
	"github.com/mo7amed4522/tayara_api/utils"
)

// AdminApi exposes staff-only operations: assignment, review, settings and FAQ management.
type AdminApi struct{}

// Assign
//
//	@Tags		Support
//	@Summary	Assign or redirect a ticket to a staff member
//	@Produce	application/json
//	@Param		id		path		int						true	"ticket id"
//	@Param		data	body		supportReq.AssignTicketReq	true	"assignment"
//	@Success	200		{object}	response.Response
//	@Router		/api/v1/support/admin/tickets/{id}/assign [post]
func (AdminApi) Assign(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithError(response.ValidationErr("invalid ticket id"), c)
		return
	}
	var req supportReq.AssignTicketReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithError(response.ValidationErr(err.Error()), c)
		return
	}
	history, err := supportService.AssignmentServices.Assign(currentActor(c), uint(id), req)
	if err != nil {
		respondErr(err, c)
		return
	}
	response.OkWithData(history, c)
}

// Assignments
//
//	@Tags		Support
//	@Summary	Assignment history for a ticket
//	@Produce	application/json
//	@Param		id	path		int	true	"ticket id"
//	@Success	200	{object}	response.Response
//	@Router		/api/v1/support/admin/tickets/{id}/assignments [get]
func (AdminApi) Assignments(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithError(response.ValidationErr("invalid ticket id"), c)
		return
	}
	rows, err := supportService.AssignmentServices.History(uint(id))
	if err != nil {
		respondErr(err, c)
		return
	}
	response.OkWithData(rows, c)
}

// Review
//
//	@Tags		Support
//	@Summary	Quality review of a ticket (append-only verdict)
//	@Produce	application/json
//	@Param		id		path		int						true	"ticket id"
//	@Param		data	body		supportReq.ReviewTicketReq	true	"verdict"
//	@Success	200		{object}	response.Response
//	@Router		/api/v1/support/admin/tickets/{id}/review [post]
func (AdminApi) Review(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithError(response.ValidationErr("invalid ticket id"), c)
		return
	}
	var req supportReq.ReviewTicketReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithError(response.ValidationErr(err.Error()), c)
		return
	}
	review, err := supportService.ReviewServices.Review(currentActor(c), uint(id), req)
	if err != nil {
		respondErr(err, c)
		return
	}
	response.OkWithData(review, c)
}

// Reviews
//
//	@Tags		Support
//	@Summary	Review trail for a ticket
//	@Produce	application/json
//	@Param		id	path		int	true	"ticket id"
//	@Success	200	{object}	response.Response
//	@Router		/api/v1/support/admin/tickets/{id}/reviews [get]
func (AdminApi) Reviews(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithError(response.ValidationErr("invalid ticket id"), c)
		return
	}
	rows, err := supportService.ReviewServices.List(uint(id))
	if err != nil {
		respondErr(err, c)
		return
	}
	response.OkWithData(rows, c)
}

// ChatSettings
//
//	@Tags		Support
//	@Summary	Read runtime chat settings (admin)
//	@Produce	application/json
//	@Success	200	{object}	response.Response
//	@Router		/api/v1/support/admin/chat-settings [get]
func (AdminApi) ChatSettings(c *gin.Context) {
	if !requireAuthority(c, constants.AuthoritySystem) {
		return
	}
	response.OkWithData(supportService.SettingsServices.List(), c)
}

// UpdateChatSettings
//
//	@Tags		Support
//	@Summary	Update runtime chat settings (admin; audited)
//	@Produce	application/json
//	@Param		data	body		supportReq.ChatSettingsUpdateReq	true	"values"
//	@Success	200		{object}	response.Response
//	@Router		/api/v1/support/admin/chat-settings [put]
func (AdminApi) UpdateChatSettings(c *gin.Context) {
	if !requireAuthority(c, constants.AuthoritySystem) {
		return
	}
	var req supportReq.ChatSettingsUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithError(response.ValidationErr(err.Error()), c)
		return
	}
	if err := supportService.SettingsServices.Update(utils.GetUserID(c), req); err != nil {
		respondErr(err, c)
		return
	}
	response.OkWithMessage("settings updated", c)
}

// ListFaqs
//
//	@Tags		Support
//	@Summary	All FAQ entries including inactive (admin)
//	@Produce	application/json
//	@Success	200	{object}	response.Response
//	@Router		/api/v1/support/admin/faqs [get]
func (AdminApi) ListFaqs(c *gin.Context) {
	rows, err := supportService.FaqServices.ListAll()
	if err != nil {
		respondErr(err, c)
		return
	}
	response.OkWithData(rows, c)
}

// UpsertFaq
//
//	@Tags		Support
//	@Summary	Create or update a FAQ entry (admin)
//	@Produce	application/json
//	@Param		data	body		supportReq.FaqReq	true	"faq"
//	@Success	200		{object}	response.Response
//	@Router		/api/v1/support/admin/faqs [post]
func (AdminApi) UpsertFaq(c *gin.Context) {
	var req supportReq.FaqReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithError(response.ValidationErr(err.Error()), c)
		return
	}
	row, err := supportService.FaqServices.Upsert(req)
	if err != nil {
		respondErr(err, c)
		return
	}
	response.OkWithData(row, c)
}

// FaqStatus
//
//	@Tags		Support
//	@Summary	Activate or deactivate a FAQ entry (admin)
//	@Produce	application/json
//	@Param		data	body		supportReq.FaqStatusReq	true	"status"
//	@Success	200		{object}	response.Response
//	@Router		/api/v1/support/admin/faqs/status [put]
func (AdminApi) FaqStatus(c *gin.Context) {
	var req supportReq.FaqStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithError(response.ValidationErr(err.Error()), c)
		return
	}
	if err := supportService.FaqServices.SetStatus(req); err != nil {
		respondErr(err, c)
		return
	}
	response.OkWithMessage("faq updated", c)
}

// requireAuthority gates admin-only handlers on the JWT authority id.
func requireAuthority(c *gin.Context, authority uint) bool {
	if utils.GetUserAuthorityId(c) != authority {
		response.FailWithError(response.ForbiddenErr("administrator only"), c)
		return false
	}
	return true
}
