package catalog

import "github.com/mo7amed4522/tayara_api/global"

// CustomTripTemplateCountry join between a custom-trip template and a supported country.
type CustomTripTemplateCountry struct {
	ID         uint `json:"id" gorm:"primarykey;comment:ID"`
	TemplateID uint `json:"templateId" form:"templateId" gorm:"index;comment:Template ID"`
	CountryID  uint `json:"countryId" form:"countryId" gorm:"index;comment:Country ID"`
	global.GVA_MODEL
}

func (CustomTripTemplateCountry) TableName() string { return "catalog_custom_trip_template_country" }
