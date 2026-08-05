package catalog

import "github.com/mo7amed4522/tayara_api/global"

// PackageActivity an activity item in a package (max ~20 per package).
type PackageActivity struct {
	ID               uint   `json:"id" gorm:"primarykey;comment:ID"`
	PackageID        uint   `json:"packageId" form:"packageId" gorm:"index;comment:Package ID"`
	Title            string `json:"title" form:"title" gorm:"size:191;comment:Activity title"`
	ImageURL         string `json:"imageURL" form:"imageURL" gorm:"size:500;comment:Activity image URL"`
	EstimatedCost    int64  `json:"estimatedCost" form:"estimatedCost" gorm:"default:0;comment:Estimated cost, minor units"`
	IncludeInPackage bool   `json:"includeInPackage" form:"includeInPackage" gorm:"default:false;comment:Included in package"`
	SortOrder        int    `json:"sortOrder" form:"sortOrder" gorm:"default:0;comment:Sort order"`
	global.GVA_MODEL
}

func (PackageActivity) TableName() string { return "catalog_package_activity" }
