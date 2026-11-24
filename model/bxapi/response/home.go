package response

import (
	"time"

	"github.com/mo7amed4522/tayara_api/model/catalog"
	"github.com/mo7amed4522/tayara_api/model/dxapi"
)

// CustomerSearchResult the unified response returned by the public search endpoint.
type CustomerSearchResult struct {
	Keyword             string                  `json:"keyword"`
	Agents              []CustomerAgentResult   `json:"agents"`
	TopExplore          []CustomerExploreResult `json:"topExplore"`
	SimilarDestinations []CustomerExploreResult `json:"similarDestinations"`
}

// CustomerAgentResult compact agent hit for search results.
type CustomerAgentResult struct {
	ID          uint    `json:"id"`
	CompanyName string  `json:"companyName"`
	HeaderImg   string  `json:"headerImg"`
	Rating      float64 `json:"rating"`      // 0–5 average, 2 decimals
	ReviewCount int     `json:"reviewCount"` // total active reviews
	Status      string  `json:"status"`
}

// CustomerPackageResult compact package hit for search results.
type CustomerPackageResult struct {
	ID              uint    `json:"id"`
	PackageCode     string  `json:"packageCode"`
	AgentID         uint    `json:"agentId"`
	CompanyName     string  `json:"companyName"`
	Name            string  `json:"name"`
	CoverImageURL   string  `json:"coverImageURL"`
	Category        string  `json:"category"`
	DaysDuration    int     `json:"daysDuration"`
	DisplayPrice    int64   `json:"displayPrice"`    // minor units
	DiscountPercent int64   `json:"discountPercent"` // basis points
	Rating          float64 `json:"rating"`
	ReviewCount     int     `json:"reviewCount"`
	Status          string  `json:"status"`
}

// CustomerExploreResult compact package result with a short description for the "top explore" list.
type CustomerExploreResult struct {
	ID              uint    `json:"id"`
	PackageCode     string  `json:"packageCode"`
	AgentID         uint    `json:"agentId"`
	CompanyName     string  `json:"companyName"`
	Name            string  `json:"name"`
	CoverImageURL   string  `json:"coverImageURL"`
	Category        string  `json:"category"`
	DaysDuration    int     `json:"daysDuration"`
	DisplayPrice    int64   `json:"displayPrice"`
	DiscountPercent int64   `json:"discountPercent"`
	Description     string  `json:"description"`
	Rating          float64 `json:"rating"`
	ReviewCount     int     `json:"reviewCount"`
	Status          string  `json:"status"`
}

// CustomerExploreTripCard a card-shaped result for the explore/trips page.
type CustomerExploreTripCard struct {
	ID              uint    `json:"id"`
	PackageCode     string  `json:"packageCode"`
	AgentID         uint    `json:"agentId"`
	CompanyName     string  `json:"companyName"`
	Name            string  `json:"name"`
	CoverImageURL   string  `json:"coverImageURL"`
	Category        string  `json:"category"`
	DaysDuration    int     `json:"daysDuration"`
	DisplayPrice    int64   `json:"displayPrice"`
	DiscountPercent int64   `json:"discountPercent"`
	Rating          float64 `json:"rating"`
	ReviewCount     int     `json:"reviewCount"`
	Status          string  `json:"status"`
}

type CustomerExploreTripsResult struct {
	Items     []CustomerExploreTripCard `json:"items"`
	Total     int64                     `json:"total"`
	TotalPage int64                     `json:"totalPage"`
}

// CustomerDestinationCard a country card for Discover Destinations.
type CustomerDestinationCard struct {
	CountryID       uint   `json:"countryId"`
	CountryName     string `json:"countryName"`
	CountryCode     string `json:"countryCode"`
	PackageCount    int    `json:"packageCount"`
	MinDisplayPrice int64  `json:"minDisplayPrice"` // minor units, 0 if none
}

// CustomerDealResult a deal card for the home deals list.
type CustomerDealResult struct {
	catalog.Package
	AgentName       string `json:"agentName"`
	AgentLogo       string `json:"agentLogo"`
	DiscountPercent int64  `json:"discountPercent"` // basis points
	ViewCount       int64  `json:"viewCount"`       // cumulative views
}

type CustomerDealsResult struct {
	Items     []CustomerDealResult `json:"items"`
	Total     int64                `json:"total"`
	TotalPage int64                `json:"totalPage"`
}

// CustomerAgentCard a top-rated agency card for the home agents list.
type CustomerAgentCard struct {
	dxapi.DxAgentProfile
	AverageRating float64 `json:"averageRating"` // 0–5
	ReviewCount   int     `json:"reviewCount"`
}

// CustomerRecommendedResult a recommended package for an authenticated user.
type CustomerRecommendedResult struct {
	catalog.Package
	AgentName   string  `json:"agentName"`
	AgentLogo   string  `json:"agentLogo"`
	Rating      float64 `json:"rating"`
	ReviewCount int     `json:"reviewCount"`
	ViewCount   int64   `json:"viewCount"`
	Wishlisted  bool    `json:"wishlisted"`
}

// AdBannerResult an active ad creative for the home banner section.
type AdBannerResult struct {
	ID              uint      `json:"id"`
	AdCode          string    `json:"adCode"`
	BannerImageURL  string    `json:"bannerImageURL"`
	OverlayTitle    string    `json:"overlayTitle"`
	OverlaySubtitle string    `json:"overlaySubtitle"`
	CtaLabel        string    `json:"ctaLabel"`
	RedirectType    string    `json:"redirectType"`
	TargetCountryID *uint     `json:"targetCountryId"`
	TargetAgentID   *uint     `json:"targetAgentId"`
	TargetPackageID *uint     `json:"targetPackageId"`
	CreatedAt       time.Time `json:"createdAt"`
}

type ActiveHomeAdsResult struct {
	Items     []AdBannerResult `json:"items"`
	Total     int64            `json:"total"`
	TotalPage int64            `json:"totalPage"`
}

// CustomerWishlistEntry a package currently wishlisted by the user.
type CustomerWishlistEntry struct {
	catalog.Package
	AgentName    string    `json:"agentName"`
	AgentLogo    string    `json:"agentLogo"`
	WishlistedAt time.Time `json:"wishlistedAt"`
}
