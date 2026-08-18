package response

import "github.com/mo7amed4522/tayara_api/model/catalog"

// PackageDetail package header + its child items (availability, hotels, activities, add-ons,
// required fields, document requirements, itinerary, includes, excludes, highlights).
type PackageDetail struct {
	catalog.Package
	Availability         []catalog.PackageAvailability        `json:"availability"`
	Hotels               []catalog.PackageHotel               `json:"hotels"`
	Activities           []catalog.PackageActivity            `json:"activities"`
	Addons               []catalog.PackageAddon               `json:"addons"`
	RequiredFields       []catalog.PackageRequiredField       `json:"requiredFields"`
	DocumentRequirements []catalog.PackageDocumentRequirement `json:"documentRequirements"`
	ItineraryDays        []catalog.ItineraryDay               `json:"itineraryDays"`
	Includes             []catalog.PackageInclusion           `json:"includes"`
	Excludes             []catalog.PackageExclusion           `json:"excludes"`
	Highlights           []catalog.PackageHighlight           `json:"highlights"`
}
