package catalog

import (
	"time"

	"github.com/mo7amed4522/tayara_api/global"
)

// PackageAvailability fixed group dates or open-calendar ranges for a package.
type PackageAvailability struct {
	ID           uint       `json:"id" gorm:"primarykey;comment:ID"`
	PackageID    uint       `json:"packageId" form:"packageId" gorm:"index;comment:Package ID"`
	StartDate    *time.Time `json:"startDate" form:"startDate" gorm:"comment:Start date"`
	EndDate      *time.Time `json:"endDate" form:"endDate" gorm:"comment:End date"`
	IsFixedGroup bool       `json:"isFixedGroup" form:"isFixedGroup" gorm:"default:false;comment:Fixed group date"`
	MaxSeats     *int       `json:"maxSeats" form:"maxSeats" gorm:"comment:Max seats (group trips)"`
	Notes        string     `json:"notes" form:"notes" gorm:"size:500;comment:Notes"`
	global.GVA_MODEL
}

func (PackageAvailability) TableName() string { return "catalog_package_availability" }
