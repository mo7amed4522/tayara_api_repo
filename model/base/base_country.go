package base

import (
	"github.com/mo7amed4522/tayara_api/model/common"
)

// Country destination/source country reference (reference table: bare ID, no soft delete).
type Country struct {
	ID                   uint           `json:"id" gorm:"primarykey;comment:ID"`
	Code                 string         `json:"code" form:"code" gorm:"size:32;uniqueIndex;comment:ISO alpha-2 code"`
	NameEn               string         `json:"nameEn" form:"nameEn" gorm:"size:128;index;comment:Name (EN)"`
	NameAr               string         `json:"nameAr" form:"nameAr" gorm:"size:128;comment:Name (AR)"`
	FlagURL              string         `json:"flagURL" form:"flagURL" gorm:"size:500;comment:Flag image URL"`
	RequiresVisaFromUAE  bool           `json:"requiresVisaFromUAE" form:"requiresVisaFromUAE" gorm:"default:false;comment:Visa required when coming from UAE"`
	IsActive             bool           `json:"isActive" form:"isActive" gorm:"default:true;comment:Is active"`
	Keywords             common.JSONMap `json:"keywords" form:"keywords" gorm:"comment:Search keywords/aliases for this country (e.g., {'keywords': ['UAE', 'United Arab Emirates', 'Emirates', 'AE']})"`
}

func (Country) TableName() string { return "base_country" }
