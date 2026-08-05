package engage

import "github.com/mo7amed4522/tayara_api/global"

// AdClick a per-click analytics record for a creative.
type AdClick struct {
	ID         uint   `json:"id" gorm:"primarykey;comment:ID"`
	AdID       uint   `json:"adId" form:"adId" gorm:"index;comment:Ad creative ID"`
	UserID     *uint  `json:"userId" form:"userId" gorm:"index;comment:Clicking user ID (nullable)"`
	SourcePage string `json:"sourcePage" form:"sourcePage" gorm:"size:64;comment:Originating page"`
	global.GVA_MODEL
}

func (AdClick) TableName() string { return "engage_ad_click" }
