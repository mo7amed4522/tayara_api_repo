package catalog

import "github.com/mo7amed4522/tayara_api/global"

// PackageHotel a hotel option for a package.
type PackageHotel struct {
	ID               uint   `json:"id" gorm:"primarykey;comment:ID"`
	PackageID        uint   `json:"packageId" form:"packageId" gorm:"index;comment:Package ID"`
	HotelStars       int    `json:"hotelStars" form:"hotelStars" gorm:"comment:Hotel stars (3/4/5)"`
	RoomType         string `json:"roomType" form:"roomType" gorm:"size:32;comment:single|double"`
	OptionName       string `json:"optionName" form:"optionName" gorm:"size:191;comment:Hotel option name"`
	PricePerNightEst int64  `json:"pricePerNightEst" form:"pricePerNightEst" gorm:"default:0;comment:Price per night estimate, minor units"`
	IncludeInPackage bool   `json:"includeInPackage" form:"includeInPackage" gorm:"default:false;comment:Included in package"`
	Nights           int    `json:"nights" form:"nights" gorm:"default:0;comment:Nights"`
	SortOrder        int    `json:"sortOrder" form:"sortOrder" gorm:"default:0;comment:Sort order"`
	global.GVA_MODEL
}

func (PackageHotel) TableName() string { return "catalog_package_hotel" }
