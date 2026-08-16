package response

import (
	"time"

	"github.com/mo7amed4522/tayara_api/model/catalog"
	"github.com/mo7amed4522/tayara_api/model/dxapi"
)

// CustomerAgentDetailResult the full agent detail page payload.
type CustomerAgentDetailResult struct {
	Agent           CustomerAgentProfile      `json:"agent"`
	Packages        []CustomerAgentPackageCard `json:"packages"`
	PackagesTotal   int64                     `json:"packagesTotal"`
	Reviews         []CustomerAgentReviewItem  `json:"reviews"`
	ReviewAvgRating float64                   `json:"reviewAvgRating"` // rounded to 2 decimals
	ReviewTotal     int                       `json:"reviewTotal"`
	TopDestination  string                    `json:"topDestination"` // most common destination country name among active packages
}

// CustomerAgentProfile the public-facing agent profile summary.
type CustomerAgentProfile struct {
	dxapi.DxAgentProfile
	AverageRating    float64 `json:"averageRating"` // 0–5, rounded to 2 decimals
	ReviewCount      int     `json:"reviewCount"`
	TrustedCount     int64   `json:"trustedCount"` // from trusted_travelers_count
	StatusLabel      string  `json:"statusLabel"` // human-readable status
}

// CustomerAgentPackageCard a package card shown on the agent detail page.
type CustomerAgentPackageCard struct {
	catalog.Package
	AgentName       string  `json:"agentName"`
	AgentLogo       string  `json:"agentLogo"`
	SalePrice       int64   `json:"salePrice"`       // displayPrice after discount, minor units
	Rating          float64 `json:"rating"`          // package-level avg rating
	ReviewCount     int     `json:"reviewCount"`
	SeatsLeft       int     `json:"seatsLeft"`       // available seats
	DepartureDate   string  `json:"departureDate"`   // formatted next departure, empty if none
	NextDepartureAt *time.Time `json:"nextDepartureAt"` // raw for the frontend
}

// CustomerAgentReviewItem a single review on the agent's profile.
type CustomerAgentReviewItem struct {
	ID          uint      `json:"id"`
	UserName    string    `json:"userName"`
	UserAvatar  string    `json:"userAvatar"` // reserved for future avatar URL
	Rating      int       `json:"rating"`
	ReviewText  string    `json:"reviewText"`
	CreatedAt   time.Time `json:"createdAt"`
}
