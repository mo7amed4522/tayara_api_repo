package base

import "github.com/mo7amed4522/tayara_api/global"

// SystemSetting runtime business configuration (commission, SLA, payment, cancellation, payout,
// currency). Seeded from the `travel` config section at init; changes are audited via the
// setting_change_history table in a later phase. Value is stored as a string and typed on read.
type SystemSetting struct {
	ID          uint   `json:"id" gorm:"primarykey;comment:ID"`
	Section     string `json:"section" form:"section" gorm:"size:64;index;comment:Settings section"`
	Key         string `json:"key" form:"key" gorm:"size:128;uniqueIndex;comment:Setting key"`
	Value       string `json:"value" form:"value" gorm:"size:255;comment:Setting value"`
	Description string `json:"description" form:"description" gorm:"size:500;comment:Description"`
	IsActive    bool   `json:"isActive" form:"isActive" gorm:"default:true;comment:Is active"`
	global.GVA_MODEL
}

func (SystemSetting) TableName() string { return "base_system_setting" }
