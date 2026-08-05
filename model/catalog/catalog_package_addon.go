package catalog

import "github.com/mo7amed4522/tayara_api/global"

// PackageAddon optional add-on a package supports (visa help, tour guide, insurance, photographer).
type PackageAddon struct {
	ID               uint   `json:"id" gorm:"primarykey;comment:ID"`
	PackageID        uint   `json:"packageId" form:"packageId" gorm:"index;comment:Package ID"`
	AddonType        string `json:"addonType" form:"addonType" gorm:"size:32;index;comment:Add-on type"`
	EstimatedPrice   int64  `json:"estimatedPrice" form:"estimatedPrice" gorm:"default:0;comment:Estimated price, minor units"`
	PriceModel       string `json:"priceModel" form:"priceModel" gorm:"size:32;comment:fixed|per_traveler"`
	IncludeInPackage bool   `json:"includeInPackage" form:"includeInPackage" gorm:"default:false;comment:Included in package"`
	Active           bool   `json:"active" form:"active" gorm:"default:true;comment:Is active"`
	global.GVA_MODEL
}

func (PackageAddon) TableName() string { return "catalog_package_addon" }
