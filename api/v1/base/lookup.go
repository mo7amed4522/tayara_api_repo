package base

import (
	"github.com/gin-gonic/gin"
	baseReq "github.com/mo7amed4522/tayara_api/model/base/request"
	commonReq "github.com/mo7amed4522/tayara_api/model/common/request"
	"github.com/mo7amed4522/tayara_api/model/common/response"
	"github.com/mo7amed4522/tayara_api/utils"
)

type LookupApi struct{}

// UpsertCategory
// @Tags Base
// @Summary Admin creates/updates a package category
// @Produce application/json
// @Param data body baseReq.PackageCategoryReq true "category payload"
// @Success 200 {object} response.Response
// @Router /api/v1/base/categories [post]
func (LookupApi) UpsertCategory(c *gin.Context) {
	var req baseReq.PackageCategoryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	cat, err := baseServiceApp.LookupServices.UpsertCategory(req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(cat, c)
}

// ListCategories
// @Tags Base
// @Summary Public list of active package categories
// @Produce application/json
// @Success 200 {object} response.Response
// @Router /api/v1/base/categories [get]
func (LookupApi) ListCategories(c *gin.Context) {
	list, err := baseServiceApp.LookupServices.ListCategories()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(list, c)
}

// DeleteCategory
// @Tags Base
// @Summary Admin deletes a package category
// @Produce application/json
// @Param query query commonReq.GetById true "category id"
// @Success 200 {object} response.Response
// @Router /api/v1/base/categories [delete]
func (LookupApi) DeleteCategory(c *gin.Context) {
	var idReq commonReq.GetById
	if err := c.ShouldBindQuery(&idReq); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := baseServiceApp.LookupServices.DeleteCategory(idReq.Uint()); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("category deleted", c)
}

// RecordSettingChange
// @Tags Base
// @Summary Admin records a system-settings edit (audit trail)
// @Produce application/json
// @Param data body baseReq.SettingChangeReq true "change payload"
// @Success 200 {object} response.Response
// @Router /api/v1/base/settings/history [post]
func (LookupApi) RecordSettingChange(c *gin.Context) {
	adminID := utils.GetUserID(c)
	var req baseReq.SettingChangeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := baseServiceApp.LookupServices.RecordChange(adminID, req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("change recorded", c)
}

// SettingHistory
// @Tags Base
// @Summary Admin views the system-settings change history
// @Produce application/json
// @Param data query baseReq.SettingHistoryPageReq true "filter"
// @Success 200 {object} response.Response
// @Router /api/v1/base/settings/history [get]
func (LookupApi) SettingHistory(c *gin.Context) {
	var req baseReq.SettingHistoryPageReq
	_ = c.ShouldBindQuery(&req)
	list, total, err := baseServiceApp.LookupServices.SettingHistory(req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(response.PageResult{List: list, Total: total, Page: req.Page, PageSize: req.PageSize}, "ok", c)
}
