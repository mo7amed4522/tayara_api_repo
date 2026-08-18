package catalog

import "github.com/mo7amed4522/tayara_api/global"

// PackageExclusion a "what's NOT included" line item.
type PackageExclusion struct {
	ID        uint   `json:"id" gorm:"primarykey;comment:ID"`
	PackageID uint   `json:"packageId" form:"packageId" gorm:"index;comment:Package ID"`
	ItemText  string `json:"itemText" form:"itemText" gorm:"size:500;comment:Exclusion description"`
	SortOrder int    `json:"sortOrder" form:"sortOrder" gorm:"default:0;comment:Sort order"`
	global.GVA_MODEL
}

func (PackageExclusion) TableName() string { return "catalog_package_exclusion" }
