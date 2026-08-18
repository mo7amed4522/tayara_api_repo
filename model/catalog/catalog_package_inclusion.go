package catalog

import "github.com/mo7amed4522/tayara_api/global"

// PackageInclusion a "what's included" line item, optionally grouped by category.
type PackageInclusion struct {
	ID        uint   `json:"id" gorm:"primarykey;comment:ID"`
	PackageID uint   `json:"packageId" form:"packageId" gorm:"index;comment:Package ID"`
	Category  string `json:"category" form:"category" gorm:"size:64;comment:Category, e.g. Accommodation / Meals / Tours"`
	ItemText  string `json:"itemText" form:"itemText" gorm:"size:500;comment:Inclusion description"`
	SortOrder int    `json:"sortOrder" form:"sortOrder" gorm:"default:0;comment:Sort order"`
	global.GVA_MODEL
}

func (PackageInclusion) TableName() string { return "catalog_package_inclusion" }
