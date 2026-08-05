package base

import (
	"github.com/mo7amed4522/tayara_api/global"
)

// SettingChangeHistory append-only audit trail for every system-settings edit (who changed
// which key from what value to what, and when). PRD mandates config changes are traceable.
type SettingChangeHistory struct {
	ID        uint   `json:"id" gorm:"primarykey;comment:ID"`
	Section   string `json:"section" form:"section" gorm:"size:64;index;comment:Settings section"`
	Key       string `json:"key" form:"key" gorm:"size:128;index;comment:Setting key"`
	OldValue  string `json:"oldValue" form:"oldValue" gorm:"size:255;comment:Previous value"`
	NewValue  string `json:"newValue" form:"newValue" gorm:"size:255;comment:New value"`
	ChangedBy *uint  `json:"changedBy" form:"changedBy" gorm:"index;comment:Acting admin user ID"`
	global.GVA_MODEL
}

func (SettingChangeHistory) TableName() string { return "base_setting_change_history" }
