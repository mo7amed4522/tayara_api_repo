package catalog

import "github.com/mo7amed4522/tayara_api/global"

// PackageRequiredField a traveler user-info field required for a package (agent-selectable + admin-locked).
type PackageRequiredField struct {
	ID                uint   `json:"id" gorm:"primarykey;comment:ID"`
	PackageID         uint   `json:"packageId" form:"packageId" gorm:"index;comment:Package ID"`
	FieldCode         string `json:"fieldCode" form:"fieldCode" gorm:"size:64;index;comment:Field code (passport_copy, dob, ...)"`
	IsAgentSelectable bool   `json:"isAgentSelectable" form:"isAgentSelectable" gorm:"default:false;comment:Agent can toggle"`
	IsAdminLocked     bool   `json:"isAdminLocked" form:"isAdminLocked" gorm:"default:false;comment:Admin-locked"`
	IsRequired        bool   `json:"isRequired" form:"isRequired" gorm:"default:true;comment:Is required"`
	SortOrder         int    `json:"sortOrder" form:"sortOrder" gorm:"default:0;comment:Sort order"`
	global.GVA_MODEL
}

func (PackageRequiredField) TableName() string { return "catalog_package_required_field" }
