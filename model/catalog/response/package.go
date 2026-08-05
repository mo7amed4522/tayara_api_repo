package response

import "github.com/mo7amed4522/tayara_api/model/catalog"

// PackageDetail package header + its child items (availability, hotels, activities, add-ons,
// required fields, document requirements).
type PackageDetail struct {
	catalog.Package
	Availability         []catalog.PackageAvailability        `json:"availability"`
	Hotels               []catalog.PackageHotel               `json:"hotels"`
	Activities           []catalog.PackageActivity            `json:"activities"`
	Addons               []catalog.PackageAddon               `json:"addons"`
	RequiredFields       []catalog.PackageRequiredField       `json:"requiredFields"`
	DocumentRequirements []catalog.PackageDocumentRequirement `json:"documentRequirements"`
}
