package catalog

import "github.com/mo7amed4522/tayara_api/global"

// CustomTripRequirementProfile the dynamic required-documents/fields per (template, country) for
// the custom-trip builder. Resolves what a traveler must submit for a given destination.
type CustomTripRequirementProfile struct {
	ID             uint   `json:"id" gorm:"primarykey;comment:ID"`
	TemplateID     uint   `json:"templateId" form:"templateId" gorm:"index;comment:Custom trip template ID"`
	CountryID      uint   `json:"countryId" form:"countryId" gorm:"index;comment:Destination country ID"`
	DocumentTypeID uint   `json:"documentTypeId" form:"documentTypeId" gorm:"index;comment:Document type ID"`
	TravelerField  string `json:"travelerField" form:"travelerField" gorm:"size:64;index;comment:Traveler field code (passport_copy, dob, ...)"`
	IsRequired     bool   `json:"isRequired" form:"isRequired" gorm:"default:true;comment:Is required"`
	Phase          string `json:"phase" form:"phase" gorm:"size:16;default:before;comment:before|after"`
	Notes          string `json:"notes" form:"notes" gorm:"size:500;comment:Notes"`
	global.GVA_MODEL
}

func (CustomTripRequirementProfile) TableName() string { return "catalog_custom_trip_requirement_profile" }
