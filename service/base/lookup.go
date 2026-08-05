package base

import (
	"errors"

	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/model/base"
	baseReq "github.com/mo7amed4522/tayara_api/model/base/request"
	"gorm.io/gorm/clause"
)

// LookupServices owns the base reference tables (package categories, settings change history).
type LookupServices struct{}

// --- package categories ---

// UpsertCategory creates or updates a package category (keyed by code).
func (s LookupServices) UpsertCategory(req baseReq.PackageCategoryReq) (*base.PackageCategory, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	cat := base.PackageCategory{
		Code: req.Code, Name: req.Name, Description: req.Description,
		SortOrder: req.SortOrder, IsActive: req.IsActive,
	}
	if err := global.GVA_DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "code"}},
		DoUpdates: clause.AssignmentColumns([]string{"name", "description", "sort_order", "is_active"}),
	}).Create(&cat).Error; err != nil {
		return nil, err
	}
	return &cat, nil
}

// ListCategories returns active package categories ordered by sort.
func (s LookupServices) ListCategories() ([]base.PackageCategory, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	var list []base.PackageCategory
	if err := global.GVA_DB.Where("is_active = ?", true).Order("sort_order asc").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// DeleteCategory soft-deletes a category (admin).
func (s LookupServices) DeleteCategory(id uint) error {
	if global.GVA_DB == nil {
		return errors.New("db not init")
	}
	return global.GVA_DB.Delete(&base.PackageCategory{}, id).Error
}

// --- settings change history ---

// RecordChange appends an audit entry for a system-settings edit.
func (s LookupServices) RecordChange(adminID uint, req baseReq.SettingChangeReq) error {
	if global.GVA_DB == nil {
		return errors.New("db not init")
	}
	by := adminID
	return global.GVA_DB.Create(&base.SettingChangeHistory{
		Section: req.Section, Key: req.Key, OldValue: req.OldValue, NewValue: req.NewValue, ChangedBy: &by,
	}).Error
}

// SettingHistory returns the change-history trail (admin), optionally filtered by key/section.
func (s LookupServices) SettingHistory(req baseReq.SettingHistoryPageReq) ([]base.SettingChangeHistory, int64, error) {
	if global.GVA_DB == nil {
		return nil, 0, errors.New("db not init")
	}
	db := global.GVA_DB.Model(&base.SettingChangeHistory{})
	if req.Key != "" {
		db = db.Where("key = ?", req.Key)
	}
	if req.Section != "" {
		db = db.Where("section = ?", req.Section)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []base.SettingChangeHistory
	if err := db.Scopes(req.Paginate()).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
