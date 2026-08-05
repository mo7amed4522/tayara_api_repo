package catalog

import (
	"time"

	"github.com/mo7amed4522/tayara_api/global"
)

// PackageView a recorded view of a package (drives the agent "Package Views" metric).
type PackageView struct {
	ID        uint       `json:"id" gorm:"primarykey;comment:ID"`
	PackageID uint       `json:"packageId" form:"packageId" gorm:"index;comment:Package ID"`
	UserID    *uint      `json:"userId" form:"userId" gorm:"index;comment:Viewing user ID (nullable)"`
	ViewDate  *time.Time `json:"viewDate" form:"viewDate" gorm:"comment:View date"`
	IP        string     `json:"ip" form:"ip" gorm:"size:64;comment:Client IP"`
	Session   string     `json:"session" form:"session" gorm:"size:128;comment:Session/source"`
	global.GVA_MODEL
}

func (PackageView) TableName() string { return "catalog_package_view" }
