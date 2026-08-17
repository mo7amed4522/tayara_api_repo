package request

import (
	"github.com/mo7amed4522/tayara_api/model/common"
	commonReq "github.com/mo7amed4522/tayara_api/model/common/request"
)

// Child item request DTOs (kept 1:1 with the catalog child entities).

type HotelReq struct {
	ID               uint   `json:"id"`
	HotelStars       int    `json:"hotelStars"`
	RoomType         string `json:"roomType"`
	OptionName       string `json:"optionName"`
	PricePerNightEst int64  `json:"pricePerNightEst"`
	IncludeInPackage bool   `json:"includeInPackage"`
	Nights           int    `json:"nights"`
	SortOrder        int    `json:"sortOrder"`
}

type ActivityReq struct {
	ID               uint   `json:"id"`
	Title            string `json:"title"`
	ImageURL         string `json:"imageURL"`
	EstimatedCost    int64  `json:"estimatedCost"`
	IncludeInPackage bool   `json:"includeInPackage"`
	SortOrder        int    `json:"sortOrder"`
}

type AddonReq struct {
	ID               uint   `json:"id"`
	AddonType        string `json:"addonType"`
	EstimatedPrice   int64  `json:"estimatedPrice"`
	PriceModel       string `json:"priceModel"`
	IncludeInPackage bool   `json:"includeInPackage"`
	Active           bool   `json:"active"`
}

type AvailabilityReq struct {
	ID           uint   `json:"id"`
	StartDate    string `json:"startDate"` // layout 2006-01-02
	EndDate      string `json:"endDate"`
	IsFixedGroup bool   `json:"isFixedGroup"`
	MaxSeats     *int   `json:"maxSeats"`
	Notes        string `json:"notes"`
}

type RequiredFieldReq struct {
	ID                uint   `json:"id"`
	FieldCode         string `json:"fieldCode"`
	IsAgentSelectable bool   `json:"isAgentSelectable"`
	IsAdminLocked     bool   `json:"isAdminLocked"`
	IsRequired        bool   `json:"isRequired"`
	SortOrder         int    `json:"sortOrder"`
}

type DocumentRequirementReq struct {
	ID             uint   `json:"id"`
	DocumentTypeID uint   `json:"documentTypeId"`
	IsRequired     bool   `json:"isRequired"`
	IsOptional     bool   `json:"isOptional"`
	RequiredAfter  string `json:"requiredAfter"`
	IsLocked       bool   `json:"isLocked"`
	Notes          string `json:"notes"`
}

// ItineraryDayReq daily itinerary request DTO.
type ItineraryDayReq struct {
	ID          uint               `json:"id"`
	DayNumber   int                `json:"dayNumber"`
	Title       string             `json:"title"`
	Description string             `json:"description"`
	SortOrder   int                `json:"sortOrder"`
	Items       []ItineraryItemReq `json:"items"`
}

// ItineraryItemReq a time-slotted entry inside an itinerary day.
type ItineraryItemReq struct {
	ID          uint   `json:"id"`
	Time        string `json:"time"`
	Title       string `json:"title"`
	Description string `json:"description"`
	SortOrder   int    `json:"sortOrder"`
}

// PackageInclusionReq a "what's included" line item.
type PackageInclusionReq struct {
	ID        uint   `json:"id"`
	Category  string `json:"category"`
	ItemText  string `json:"itemText"`
	SortOrder int    `json:"sortOrder"`
}

// PackageExclusionReq a "what's NOT included" line item.
type PackageExclusionReq struct {
	ID        uint   `json:"id"`
	ItemText  string `json:"itemText"`
	SortOrder int    `json:"sortOrder"`
}

// PackageHighlightReq a highlight/amenity tag for a package.
type PackageHighlightReq struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	IconURL   string `json:"iconURL"`
	SortOrder int    `json:"sortOrder"`
}

// PackageReq create/update payload for a package + its child items.
type PackageReq struct {
	ID                   uint                     `json:"id"` // 0 on create
	Type                 string                   `json:"type"`
	Name                 string                   `json:"name"`
	CoverImageURL        string                   `json:"coverImageURL"`
	Description          string                   `json:"description"`
	GalleryImages        common.JSONMap           `json:"galleryImages"`
	DestinationCountryID uint                     `json:"destinationCountryId"`
	Location             string                   `json:"location"`
	Category             string                   `json:"category"`
	DaysDuration         int                      `json:"daysDuration"`
	AvailabilityType     string                   `json:"availabilityType"`
	DisplayPrice         int64                    `json:"displayPrice"`
	EstimatedPrice       int64                    `json:"estimatedPrice"`
	DiscountPercent      int64                    `json:"discountPercent"`
	// Tour detail fields.
	StatusTour    string `json:"statusTour"`
	TimeDisplay   string `json:"timeDisplay"`
	Destination   string `json:"destination"`
	OriginalPrice int64  `json:"originalPrice"`
	PromoTag      string `json:"promoTag"`
	// Child items.
	Availability         []AvailabilityReq        `json:"availability"`
	Hotels               []HotelReq               `json:"hotels"`
	Activities           []ActivityReq            `json:"activities"`
	Addons               []AddonReq               `json:"addons"`
	RequiredFields       []RequiredFieldReq       `json:"requiredFields"`
	DocumentRequirements []DocumentRequirementReq `json:"documentRequirements"`
	ItineraryDays        []ItineraryDayReq        `json:"itineraryDays"`
	Includes             []PackageInclusionReq    `json:"includes"`
	Excludes             []PackageExclusionReq    `json:"excludes"`
	Highlights           []PackageHighlightReq    `json:"highlights"`
}

// PackagePageReq paginated listing with optional status filter.
type PackagePageReq struct {
	commonReq.PageInfo
	Status string `json:"status" form:"status"`
	Type   string `json:"type" form:"type"`
}

// PackageAdminStatusReq admin moderation action on a package.
type PackageAdminStatusReq struct {
	ID            uint   `json:"id"`
	Status        string `json:"status"` // active | hidden | flagged | disabled | pending_approval
	Note          string `json:"note"`
	FlagIssueType string `json:"flagIssueType"`
	FlagDetails   string `json:"flagDetails"`
}
