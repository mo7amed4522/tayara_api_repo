package engage

import (
	"github.com/gin-gonic/gin"
	commonReq "github.com/mo7amed4522/tayara_api/model/common/request"
	"github.com/mo7amed4522/tayara_api/model/common/response"
	engageReq "github.com/mo7amed4522/tayara_api/model/engage/request"
	"github.com/mo7amed4522/tayara_api/utils"
)

type NotificationApi struct{}

// ListNotifications
//	@Tags		Engage
//	@Summary	Admin lists notification log (all recipients, filterable)
//	@Produce	application/json
//	@Param		data	query		engageReq.NotificationPageReq	true	"filter"
//	@Success	200		{object}	response.Response
//	@Router		/api/v1/engage/admin/notifications [get]
func (NotificationApi) ListNotifications(c *gin.Context) {
	var req engageReq.NotificationPageReq
	_ = c.ShouldBindQuery(&req)
	list, total, totalPage, err := engageService.NotificationServices.ListNotifications(req, "", 0)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(response.PageResult{List: list, Total: total, TotalPage: totalPage, Page: req.Page, PageSize: req.PageSize}, "ok", c)
}

// ListMyNotifications
//	@Tags		Engage
//	@Summary	User lists their own notifications
//	@Produce	application/json
//	@Param		data	query		engageReq.NotificationPageReq	true	"filter"
//	@Success	200		{object}	response.Response
//	@Router		/api/v1/engage/notifications [get]
func (NotificationApi) ListMyNotifications(c *gin.Context) {
	userID := utils.GetBxUserID(c)
	var req engageReq.NotificationPageReq
	_ = c.ShouldBindQuery(&req)
	list, total, totalPage, err := engageService.NotificationServices.ListNotifications(req, "user", userID)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(response.PageResult{List: list, Total: total, TotalPage: totalPage, Page: req.Page, PageSize: req.PageSize}, "ok", c)
}

// Resend
//	@Tags		Engage
//	@Summary	Admin resends a notification
//	@Produce	application/json
//	@Param		data	body		commonReq.GetById	true	"notification id"
//	@Success	200		{object}	response.Response
//	@Router		/api/v1/engage/admin/notifications/resend [post]
func (NotificationApi) Resend(c *gin.Context) {
	var req commonReq.GetById
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	rec, err := engageService.NotificationServices.Resend(uint(req.ID))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(rec, c)
}

// UpsertTemplate
//	@Tags		Engage
//	@Summary	Admin creates/updates a notification template
//	@Produce	application/json
//	@Param		data	body		engageReq.NotificationTemplateReq	true	"template payload"
//	@Success	200		{object}	response.Response
//	@Router		/api/v1/engage/admin/notification-templates [post]
func (NotificationApi) UpsertTemplate(c *gin.Context) {
	var req engageReq.NotificationTemplateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	tpl, err := engageService.NotificationServices.UpsertTemplate(req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(tpl, c)
}

// ListTemplates
//	@Tags		Engage
//	@Summary	Admin lists all notification templates (design/moderate)
//	@Produce	application/json
//	@Success	200	{object}	response.Response
//	@Router		/api/v1/engage/admin/notification-templates [get]
func (NotificationApi) ListTemplates(c *gin.Context) {
	list, err := engageService.NotificationServices.ListTemplates()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(list, c)
}

// DeleteTemplate
//	@Tags		Engage
//	@Summary	Admin deletes a notification template
//	@Produce	application/json
//	@Param		query	query		commonReq.GetById	true	"template id"
//	@Success	200		{object}	response.Response
//	@Router		/api/v1/engage/admin/notification-templates [delete]
func (NotificationApi) DeleteTemplate(c *gin.Context) {
	var idReq commonReq.GetById
	if err := c.ShouldBindQuery(&idReq); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := engageService.NotificationServices.DeleteTemplate(idReq.Uint()); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("template deleted", c)
}
