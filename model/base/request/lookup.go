package request

import (
	commonReq "github.com/mo7amed4522/tayara_api/model/common/request"
)

// PackageCategoryReq create/update a package category (keyed by code).
type PackageCategoryReq struct {
	Code        string `json:"code" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	SortOrder   int    `json:"sortOrder"`
	IsActive    bool   `json:"isActive"`
}

// SettingChangeReq an auditable system-settings edit (recorded to change history).
type SettingChangeReq struct {
	Section  string `json:"section" binding:"required"`
	Key      string `json:"key" binding:"required"`
	OldValue string `json:"oldValue"`
	NewValue string `json:"newValue" binding:"required"`
}

// SettingUpdateReq upsert a single system setting by key (typed on read by callers).
type SettingUpdateReq struct {
	Section     string `json:"section"` // leave empty to keep the existing section
	Key         string `json:"key" binding:"required"`
	Value       string `json:"value" binding:"required"`
	Description string `json:"description"`
}

// SettingHistoryPageReq paginated change-history listing with optional key/section filter.
type SettingHistoryPageReq struct {
	commonReq.PageInfo
	Key     string `json:"key" form:"key"`
	Section string `json:"section" form:"section"`
}
