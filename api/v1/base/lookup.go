package base

import (
	"context"
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/mo7amed4522/tayara_api/constants"
	baseReq "github.com/mo7amed4522/tayara_api/model/base/request"
	commonReq "github.com/mo7amed4522/tayara_api/model/common/request"
	"github.com/mo7amed4522/tayara_api/model/common/response"
	"github.com/mo7amed4522/tayara_api/utils"
	"gorm.io/gorm"
)

type LookupApi struct{}

// requireSuperAdmin returns 403 if the caller is not a platform super admin (authorityId == 1).
// All settings management endpoints are super-admin-only; non-super-admin system users (agents, etc.)
// are rejected before any DB access occurs.
func requireSuperAdmin(c *gin.Context) bool {
	authorityID := utils.GetUserAuthorityId(c)
	if authorityID != constants.AuthoritySystem {
		response.FailWithMessage("forbidden: super-admin only", c)
		c.Abort()
		return false
	}
	return true
}

// UpsertCategory
//
//	@Tags		Base
//	@Summary	Admin creates/updates a package category
//	@Produce	application/json
//	@Param		data	body		baseReq.PackageCategoryReq	true	"category payload"
//	@Success	200		{object}	response.Response
//	@Router		/api/v1/base/categories [post]
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
//
//	@Tags		Base
//	@Summary	Public list of active package categories
//	@Produce	application/json
//	@Param		page		query		int	false	"page number"	default(1)
//	@Param		pageSize	query		int	false	"page size"	default(10)
//	@Success	200	{object}	response.Response
//	@Router		/api/v1/base/categories [get]
func (LookupApi) ListCategories(c *gin.Context) {
	var pageInfo commonReq.PageInfo
	if err := c.ShouldBindQuery(&pageInfo); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, totalPage, err := baseServiceApp.LookupServices.ListCategories(pageInfo.Page, pageInfo.PageSize)
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

// DeleteCategory
//
//	@Tags		Base
//	@Summary	Admin deletes a package category
//	@Produce	application/json
//	@Param		query	query		commonReq.GetById	true	"category id"
//	@Success	200		{object}	response.Response
//	@Router		/api/v1/base/categories [delete]
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
//
//	@Tags		Base
//	@Summary	Admin records a system-settings edit (audit trail)
//	@Produce	application/json
//	@Param		data	body		baseReq.SettingChangeReq	true	"change payload"
//	@Success	200		{object}	response.Response
//	@Router		/api/v1/base/settings/history [post]
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
//
//	@Tags		Base
//	@Summary	Admin views the system-settings change history
//	@Produce	application/json
//	@Param		data	query		baseReq.SettingHistoryPageReq	true	"filter"
//	@Success	200		{object}	response.Response
//	@Router		/api/v1/base/settings/history [get]
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
//
//	@Tags		Base
//	@Summary	Super-admin views all system settings grouped by section
//	@Produce	application/json
//	@Param		page		query		int	false	"page number"	default(1)
//	@Param		pageSize	query		int	false	"page size"	default(10)
//	@Success	200	{object}	response.Response
//	@Router		/api/v1/base/settings [get]
func (LookupApi) Settings(c *gin.Context) {
	if !requireSuperAdmin(c) {
		return
	}
	var pageInfo commonReq.PageInfo
	if err := c.ShouldBindQuery(&pageInfo); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	result, err := baseServiceApp.LookupServices.SettingsBySection(pageInfo.Page, pageInfo.PageSize)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(response.PageResult{
		List: result.Data, Total: result.Total, TotalPage: result.TotalPage,
		Page: result.Page, PageSize: result.PageSize,
	}, "ok", c)
}

// GetSetting
//
//	@Tags		Base
//	@Summary	Super-admin retrieves a single system setting by key
//	@Produce	application/json
//	@Param		key	path		string	true	"setting key"
//	@Success	200	{object}	response.Response
//	@Router		/api/v1/base/settings/{key} [get]
func (LookupApi) GetSetting(c *gin.Context) {
	if !requireSuperAdmin(c) {
		return
	}
	key := c.Param("key")
	if key == "" {
		response.FailWithMessage("key required", c)
		return
	}
	setting, err := baseServiceApp.LookupServices.GetSetting(key)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.FailWithMessage("setting not found", c)
			return
		}
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(setting, c)
}

// UpdateSetting
//
//	@Tags		Base
//	@Summary	Super-admin updates or creates a system setting (audited)
//	@Produce	application/json
//	@Param		data	body		baseReq.SettingUpdateReq	true	"setting payload"
//	@Success	200		{object}	response.Response
//	@Router		/api/v1/base/settings [put]
func (LookupApi) UpdateSetting(c *gin.Context) {
	if !requireSuperAdmin(c) {
		return
	}
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

// DeleteSetting
//
//	@Tags		Base
//	@Summary	Super-admin deletes a system setting override (reverts to config default)
//	@Produce	application/json
//	@Param		key	path		string	true	"setting key"
//	@Success	200	{object}	response.Response
//	@Router		/api/v1/base/settings/{key} [delete]
func (LookupApi) DeleteSetting(c *gin.Context) {
	if !requireSuperAdmin(c) {
		return
	}
	key := c.Param("key")
	if key == "" {
		response.FailWithMessage("key required", c)
		return
	}
	if err := baseServiceApp.LookupServices.DeleteSetting(key); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.FailWithMessage("setting not found", c)
			return
		}
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("setting deleted", c)
}

// ConfigKeys
//
//	@Tags		Base
//	@Summary	Super-admin lists every config.yaml key that can be edited at runtime
//	@Produce	application/json
//	@Param		page		query		int	false	"page number"	default(1)
//	@Param		pageSize	query		int	false	"page size"	default(10)
//	@Success	200	{object}	response.Response
//	@Router		/api/v1/base/settings/keys [get]
func (LookupApi) ConfigKeys(c *gin.Context) {
	if !requireSuperAdmin(c) {
		return
	}
	var pageInfo commonReq.PageInfo
	if err := c.ShouldBindQuery(&pageInfo); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	result, err := baseServiceApp.LookupServices.ListAllConfigKeys(pageInfo.Page, pageInfo.PageSize)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(response.PageResult{
		List: result.Data, Total: result.Total, TotalPage: result.TotalPage,
		Page: result.Page, PageSize: result.PageSize,
	}, "ok", c)
}

// ResetSection
//
//	@Tags		Base
//	@Summary	Super-admin resets a settings section to its seeded defaults
//	@Produce	application/json
//	@Param		section	query		string	true	"settings section"
//	@Success	200		{object}	response.Response
//	@Router		/api/v1/base/settings/reset [post]
func (LookupApi) ResetSection(c *gin.Context) {
	if !requireSuperAdmin(c) {
		return
	}
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
