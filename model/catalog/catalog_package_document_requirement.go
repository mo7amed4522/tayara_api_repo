package catalog

import "github.com/mo7amed4522/tayara_api/global"

// PackageDocumentRequirement maps the document library to a package (required after booking/payment).
type PackageDocumentRequirement struct {
	ID             uint   `json:"id" gorm:"primarykey;comment:ID"`
	PackageID      uint   `json:"packageId" form:"packageId" gorm:"index;comment:Package ID"`
	DocumentTypeID uint   `json:"documentTypeId" form:"documentTypeId" gorm:"index;comment:Document type ID"`
	IsRequired     bool   `json:"isRequired" form:"isRequired" gorm:"default:true;comment:Required"`
	IsOptional     bool   `json:"isOptional" form:"isOptional" gorm:"default:false;comment:Optional"`
	RequiredAfter  string `json:"requiredAfter" form:"requiredAfter" gorm:"size:32;default:payment;comment:booking|payment"`
	IsLocked       bool   `json:"isLocked" form:"isLocked" gorm:"default:false;comment:Locked once price confirmed"`
	Notes          string `json:"notes" form:"notes" gorm:"size:500;comment:Notes"`
	global.GVA_MODEL
}

func (PackageDocumentRequirement) TableName() string { return "catalog_package_document_requirement" }
