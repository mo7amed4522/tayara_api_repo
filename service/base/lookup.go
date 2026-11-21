package base

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/mo7amed4522/tayara_api/constants"
	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/model/base"
	baseReq "github.com/mo7amed4522/tayara_api/model/base/request"
	"github.com/mo7amed4522/tayara_api/utils"
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

// ListCategories returns a paginated list of active package categories.
func (s LookupServices) ListCategories(page, pageSize int) ([]base.PackageCategory, int64, int64, error) {
	if global.GVA_DB == nil {
		return nil, 0, 0, errors.New("db not init")
	}
	var list []base.PackageCategory
	countDB := global.GVA_DB.Model(&base.PackageCategory{}).Where("is_active = ?", true)
	queryDB := global.GVA_DB.Model(&base.PackageCategory{}).
		Where("is_active = ?", true).
		Order("sort_order asc")
	pr, err := utils.PaginateWithCustomQuery(countDB, queryDB, page, pageSize, &list)
	if err != nil {
		return nil, 0, 0, err
	}
	return list, pr.Total, pr.TotalPage, nil
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
func (s LookupServices) SettingHistory(req baseReq.SettingHistoryPageReq) ([]base.SettingChangeHistory, int64, int64, error) {
	if global.GVA_DB == nil {
		return nil, 0, 0, errors.New("db not init")
	}
	db := global.GVA_DB.Model(&base.SettingChangeHistory{})
	if req.Key != "" {
		db = db.Where("key = ?", req.Key)
	}
	if req.Section != "" {
		db = db.Where("section = ?", req.Section)
	}
	var list []base.SettingChangeHistory
	pr, err := utils.PaginateWithCustomQuery(db, db.Order("created_at desc"), req.Page, req.PageSize, &list)
	if err != nil {
		return nil, 0, 0, err
	}
	return list, pr.Total, pr.TotalPage, nil
}

// --- system settings management (PRD §6.10 System Settings page) ---

// SettingsBySection returns a paginated settings list grouped by section.
func (s LookupServices) SettingsBySection(page, pageSize int) (*utils.PaginateResult, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	var rows []base.SystemSetting
	countDB := global.GVA_DB.Model(&base.SystemSetting{}).Where("is_active = ?", true)
	queryDB := global.GVA_DB.Model(&base.SystemSetting{}).
		Where("is_active = ?", true).
		Order("id asc")
	result, err := utils.PaginateWithCustomQuery(countDB, queryDB, page, pageSize, &rows)
	if err != nil {
		return nil, err
	}
	out := make(map[string][]base.SystemSetting)
	for _, r := range rows {
		out[r.Section] = append(out[r.Section], r)
	}
	result.Data = out
	return result, nil
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

// --- config key enumeration ---

// ConfigKeyValue represents a single editable config key with its current value and section.
type ConfigKeyValue struct {
	Section string `json:"section"`
	Key     string `json:"key"`
	Value   string `json:"value"`
}

// ListAllConfigKeys enumerates every leaf key in config.yaml and returns them with their current
// values (DB overrides take precedence over the YAML default). The returned list is sorted by
// section then key for stable frontend rendering.
func (s LookupServices) ListAllConfigKeys(page, pageSize int) (*utils.PaginateResult, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	// Build a quick lookup of DB overrides keyed by "section.key".
	dbOverrides := make(map[string]string)
	var rows []base.SystemSetting
	if err := global.GVA_DB.Where("is_active = ?", true).Find(&rows).Error; err != nil {
		return nil, err
	}
	for _, r := range rows {
		dbOverrides[r.Section+"."+r.Key] = r.Value
	}

	// Walk the config struct via reflection and collect every leaf.
	parsed := global.GVA_CONFIG
	leaves := reflectValueTree("", "", parsed)

	// Merge with DB overrides and build the result.
	out := make([]ConfigKeyValue, 0, len(leaves))
	for _, leaf := range leaves {
		composite := leaf.section + "." + leaf.key
		val := leaf.value
		if override, ok := dbOverrides[composite]; ok {
			val = override
		}
		out = append(out, ConfigKeyValue{Section: leaf.section, Key: leaf.key, Value: val})
	}
	// Stable sort by section then key.
	sortConfigKeys(out)

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	} else if pageSize > 100 {
		pageSize = 100
	}
	total := int64(len(out))
	totalPage := (total + int64(pageSize) - 1) / int64(pageSize)
	start := len(out)
	if totalPage > 0 && int64(page) <= totalPage {
		start = (page - 1) * pageSize
	}
	end := start + pageSize
	if end > len(out) {
		end = len(out)
	}
	return &utils.PaginateResult{
		Total: total, Page: page, PageSize: pageSize, TotalPage: totalPage,
		Data: out[start:end],
	}, nil
}

type leaf struct {
	section string
	key     string
	value   string
}

// reflectValueTree walks a struct and produces every leaf key-path as a section+key pair.
// topSection is the config top-level key (e.g. "jwt", "travel"); topKey is unused at the top
// level and only populated for nested structs inside a section.
func reflectValueTree(topSection, topKey string, v interface{}) []leaf {
	var out []leaf
	rv := reflect.ValueOf(v)
	// Dereference pointer if needed.
	for rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return out
	}
	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		val := rv.Field(i)

		// Skip unexported fields.
		if field.PkgPath != "" {
			continue
		}

		tag := field.Tag.Get("yaml")
		if tag == "" || tag == "-" {
			continue
		}
		keyName := tag

		// Build the path segment.
		var pathSeg string
		if topSection != "" {
			pathSeg = topSection + "." + keyName
		} else {
			pathSeg = keyName
		}

		// Handle slices of structs (e.g. RedisList, DBList) — flatten each element under the same key prefix.
		if val.Kind() == reflect.Slice {
			for j := 0; j < val.Len(); j++ {
				elem := val.Index(j)
				for elem.Kind() == reflect.Ptr {
					elem = elem.Elem()
				}
				if elem.Kind() == reflect.Struct {
					sectionPrefix := pathSeg
					if topKey != "" {
						sectionPrefix = topSection + "." + topKey
					}
					out = append(out, reflectValueTree(sectionPrefix, keyName, elem.Interface())...)
				} else {
					out = append(out, leaf{
						section: pathSeg,
						key:     fmt.Sprintf("[%d]", j),
						value:   formatValue(elem.Interface()),
					})
				}
			}
			continue
		}

		// Recurse into nested structs.
		if val.Kind() == reflect.Ptr {
			if val.IsNil() {
				continue
			}
			val = val.Elem()
		}
		if val.Kind() == reflect.Struct {
			// Determine the effective section for nested structs.
			section := pathSeg
			if topSection == "" {
				// Top-level struct field: the field name itself becomes the section.
				section = keyName
			}
			out = append(out, reflectValueTree(section, keyName, val.Interface())...)
			continue
		}

		// Leaf primitive value.
		out = append(out, leaf{
			section: pathSeg,
			key:     keyName,
			value:   formatValue(val.Interface()),
		})
	}
	return out
}

// formatValue converts a Go value to its string representation for storage in the settings table.
func formatValue(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case bool:
		return strconv.FormatBool(val)
	case int:
		return strconv.Itoa(val)
	case int8:
		return strconv.FormatInt(int64(val), 10)
	case int16:
		return strconv.FormatInt(int64(val), 10)
	case int32:
		return strconv.FormatInt(int64(val), 10)
	case int64:
		return strconv.FormatInt(val, 10)
	case uint:
		return strconv.FormatUint(uint64(val), 10)
	case uint8:
		return strconv.FormatUint(uint64(val), 10)
	case uint16:
		return strconv.FormatUint(uint64(val), 10)
	case uint32:
		return strconv.FormatUint(uint64(val), 10)
	case uint64:
		return strconv.FormatUint(val, 10)
	case float32:
		return strconv.FormatFloat(float64(val), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	case []byte:
		return string(val)
	case nil:
		return ""
	default:
		// Fallback: marshal to JSON.
		b, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("%v", v)
		}
		return string(b)
	}
}

// sortConfigKeys sorts in-place by section ascending, then key ascending.
func sortConfigKeys(keys []ConfigKeyValue) {
	for i := 1; i < len(keys); i++ {
		for j := i; j > 0 && (keys[j].Section < keys[j-1].Section || (keys[j].Section == keys[j-1].Section && keys[j].Key < keys[j-1].Key)); j-- {
			keys[j], keys[j-1] = keys[j-1], keys[j]
		}
	}
}

// DeleteSetting removes a setting override from the database, causing the system to fall back to
// the config.yaml default on the next read. Soft-delete (set is_active=false) preserves audit
// history; hard-delete is NOT used so that change-history records remain valid.
func (s LookupServices) DeleteSetting(key string) error {
	if global.GVA_DB == nil {
		return errors.New("db not init")
	}
	result := global.GVA_DB.Model(&base.SystemSetting{}).Where("key = ?", key).Update("is_active", false)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
