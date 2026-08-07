package base

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/mo7amed4522/tayara_api/constants"
	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/model/base"
	baseReq "github.com/mo7amed4522/tayara_api/model/base/request"
	"gorm.io/gorm"
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

// SettingHistory returns the settings history trail (admin), optionally filtered by key/section.
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

// --- system settings management (PRD §6.10 System Settings page) ---

// SettingsBySection returns all settings, grouped by section (the System Settings page contract).
func (s LookupServices) SettingsBySection() (map[string][]base.SystemSetting, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	var rows []base.SystemSetting
	if err := global.GVA_DB.Where("is_active = ?", true).Order("id asc").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make(map[string][]base.SystemSetting)
	for _, r := range rows {
		out[r.Section] = append(out[r.Section], r)
	}
	return out, nil
}

// GetSetting returns a single setting by key (used by business logic to read typed values).
func (s LookupServices) GetSetting(key string) (*base.SystemSetting, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	var st base.SystemSetting
	if err := global.GVA_DB.Where("key = ?", key).First(&st).Error; err != nil {
		return nil, err
	}
	return &st, nil
}

// UpdateSetting upserts a setting value and records the change in the audit trail. The change is
// captured before update so the audit row always holds the real old→new pair.
func (s LookupServices) UpdateSetting(adminID *uint, req baseReq.SettingUpdateReq) (*base.SystemSetting, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	var st base.SystemSetting
	err := global.GVA_DB.Where("key = ?", req.Key).First(&st).Error
	isNew := errors.Is(err, gorm.ErrRecordNotFound)
	if err != nil && !isNew {
		return nil, err
	}
	oldValue := st.Value
	newValue := req.Value
	if req.Section == "" {
		req.Section = st.Section
	}
	if err := global.GVA_DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"section", "value", "description", "is_active"}),
	}).Create(&base.SystemSetting{
		Section: req.Section, Key: req.Key, Value: newValue,
		Description: req.Description, IsActive: true,
	}).Error; err != nil {
		return nil, err
	}
	// audit the change (skip audit on first insert — no old value to restore from)
	if !isNew && oldValue != newValue {
		_ = global.GVA_DB.Create(&base.SettingChangeHistory{
			Section: req.Section, Key: req.Key, OldValue: oldValue, NewValue: newValue, ChangedBy: adminID,
		}).Error
	}
	updated := &base.SystemSetting{Section: req.Section, Key: req.Key, Value: newValue, Description: req.Description}
	return updated, nil
}

// ResetSection restores every setting in a section to its seed default (seeded from the
// `travel` config at init). Audit records the bulk reset so it is traceable.
// seedDefaultsForSection returns the deploy-time default value (from the `travel` config) for
// every setting key in a section. Used by ResetSection to restore a section to its seeded state.
func seedDefaultsForSection(section string) map[string]string {
	t := global.GVA_CONFIG.Travel
	switch section {
	case constants.SettingSectionCurrency:
		return map[string]string{
			constants.SettingKeyDefaultCurrency:     t.DefaultCurrency,
			constants.SettingKeySupportedCurrencies: strings.Join(t.SupportedCurrencies, ","),
		}
	case constants.SettingSectionCommission:
		return map[string]string{
			"basic_plan_percent":                 strconv.FormatInt(t.Commission.BasicPlanPercent, 10),
			"premium_plan_percent":               strconv.FormatInt(t.Commission.PremiumPlanPercent, 10),
			"allow_manual_override":              strconv.FormatBool(t.Commission.AllowManualOverride),
			constants.SettingKeyStripeFeePercent: "230",
		}
	case constants.SettingSectionSLA:
		return map[string]string{
			"agent_response_hours":    strconv.Itoa(t.SLA.AgentResponseHours),
			"user_payment_hours":      strconv.Itoa(t.SLA.UserPaymentHours),
			"warning_threshold_hours": strconv.Itoa(t.SLA.WarningThresholdHours),
			"auto_alert_on_breach":    strconv.FormatBool(t.SLA.AutoAlertOnBreach),
		}
	case constants.SettingSectionPayment:
		return map[string]string{
			"hold_duration_hours":        strconv.Itoa(t.Payment.HoldDurationHours),
			"auto_capture":               strconv.FormatBool(t.Payment.AutoCapture),
			"auto_refund_before_capture": strconv.FormatBool(t.Payment.AutoRefundBeforeCapture),
			"partial_refund_enabled":     strconv.FormatBool(t.Payment.PartialRefundEnabled),
		}
	case constants.SettingSectionCancellation:
		return map[string]string{
			"auto_cancel_price_pending_hours": strconv.Itoa(t.Cancellation.AutoCancelPricePendingHours),
			"auto_cancel_kyc_pending_hours":   strconv.Itoa(t.Cancellation.AutoCancelKYCPendingHours),
			"allow_admin_force_cancel":        strconv.FormatBool(t.Cancellation.AllowAdminForceCancel),
		}
	case constants.SettingSectionPayout:
		return map[string]string{
			"manual_only":                strconv.FormatBool(t.Payout.ManualOnly),
			"require_admin_verification": strconv.FormatBool(t.Payout.RequireAdminVerification),
			"require_confirmation_file":  strconv.FormatBool(t.Payout.RequireConfirmationFile),
		}
	default:
		return nil
	}
}

func (s LookupServices) ResetSection(ctx context.Context, adminID uint, section string) (int, error) {
	if global.GVA_DB == nil {
		return 0, errors.New("db not init")
	}
	// resolve seed defaults for the section from the config
	defaults := seedDefaultsForSection(section)
	if len(defaults) == 0 {
		return 0, errors.New("unknown settings section")
	}
	updated := 0
	err := global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		for key, val := range defaults {
			var st base.SystemSetting
			err := tx.Where("key = ?", key).First(&st).Error
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			if st.Value != val {
				if err := tx.Create(&base.SettingChangeHistory{
					Section: section, Key: key, OldValue: st.Value, NewValue: val, ChangedBy: &adminID,
				}).Error; err != nil {
					return err
				}
				updated++
			}
			if err := tx.Model(&base.SystemSetting{}).Where("key = ?", key).
				Updates(map[string]interface{}{"value": val, "section": section}).Error; err != nil {
				return err
			}
		}
		return nil
	})
	return updated, err
}
