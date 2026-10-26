package base

import (
	"github.com/mo7amed4522/tayara_api/model/common"
)

// City cities within a country (for custom-trip templates, ads, package destinations).
type City struct {
	ID        uint           `json:"id" gorm:"primarykey;comment:ID"`
	CountryID uint           `json:"countryId" form:"countryId" gorm:"index;comment:Country ID"`
	Name      string         `json:"name" form:"name" gorm:"size:128;index;comment:City name"`
	IsActive  bool           `json:"isActive" form:"isActive" gorm:"default:true;comment:Is active"`
	Keywords  common.JSONMap `json:"keywords" form:"keywords" gorm:"comment:Search keywords/aliases for this city (e.g., {'keywords': ['Dubai', 'DXB']})"`
}

func (City) TableName() string { return "base_city" }
