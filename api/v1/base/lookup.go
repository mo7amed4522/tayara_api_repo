package base

import (
	"context"

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
	list, total, totalPage, err := baseServiceApp.LookupServices.SettingHistory(req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(response.PageResult{List: list, Total: total, TotalPage: totalPage, Page: req.Page, PageSize: req.PageSize}, "ok", c)
}

// Settings
// @Tags Base
// @Summary Admin views all system settings grouped by section
// @Produce application/json
// @Success 200 {object} response.Response
// @Router /api/v1/base/settings [get]
func (LookupApi) Settings(c *gin.Context) {
	sections, err := baseServiceApp.LookupServices.SettingsBySection()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(sections, c)
}

// UpdateSetting
// @Tags Base
// @Summary Admin updates a system setting (audited)
// @Produce application/json
// @Param data body baseReq.SettingUpdateReq true "setting payload"
// @Success 200 {object} response.Response
// @Router /api/v1/base/settings [put]
func (LookupApi) UpdateSetting(c *gin.Context) {
	adminID := utils.GetUserID(c)
	var req baseReq.SettingUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	by := adminID
	updated, err := baseServiceApp.LookupServices.UpdateSetting(&by, req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(updated, c)
}

// ResetSection
// @Tags Base
// @Summary Admin resets a settings section to its seeded defaults
// @Produce application/json
// @Param section query string true "settings section"
// @Success 200 {object} response.Response
// @Router /api/v1/base/settings/reset [post]
func (LookupApi) ResetSection(c *gin.Context) {
	adminID := utils.GetUserID(c)
	section := c.Query("section")
	if section == "" {
		response.FailWithMessage("section required", c)
		return
	}
	updated, err := baseServiceApp.LookupServices.ResetSection(context.Background(), adminID, section)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(gin.H{"section": section, "resetCount": updated}, c)
}
