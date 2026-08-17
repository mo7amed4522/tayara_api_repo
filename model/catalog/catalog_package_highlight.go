package catalog

import "github.com/mo7amed4522/tayara_api/global"

// PackageHighlight a highlight feature / amenity tag for a package (e.g. Full Board, Expert Guide, 5 Star Ship).
type PackageHighlight struct {
	ID        uint   `json:"id" gorm:"primarykey;comment:ID"`
	PackageID uint   `json:"packageId" form:"packageId" gorm:"index;comment:Package ID"`
	Name      string `json:"name" form:"name" gorm:"size:128;comment:Highlight label"`
	IconURL   string `json:"iconURL" form:"iconURL" gorm:"size:500;comment:Optional icon/illustration URL"`
	SortOrder int    `json:"sortOrder" form:"sortOrder" gorm:"default:0;comment:Sort order"`
	global.GVA_MODEL
}

func (PackageHighlight) TableName() string { return "catalog_package_highlight" }
