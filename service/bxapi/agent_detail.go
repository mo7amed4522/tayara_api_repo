package bxapi

import (
	"errors"
	"fmt"
	"strings"

	"github.com/mo7amed4522/tayara_api/constants"
	"github.com/mo7amed4522/tayara_api/global"
	bxapiModel "github.com/mo7amed4522/tayara_api/model/bxapi"
	bxapiReq "github.com/mo7amed4522/tayara_api/model/bxapi/request"
	bxapiResp "github.com/mo7amed4522/tayara_api/model/bxapi/response"
	"github.com/mo7amed4522/tayara_api/model/catalog"
	"github.com/mo7amed4522/tayara_api/model/dxapi"
	"github.com/mo7amed4522/tayara_api/model/engage"
	"github.com/mo7amed4522/tayara_api/utils"
)

// AgentDetailServices owns the customer-facing agent detail aggregations.
type AgentDetailServices struct{}

// ---------------------------------------------------------------------------
// Agent detail page
// ---------------------------------------------------------------------------

// CustomerAgentDetail returns the full agent profile page payload:
// agent summary (rating, reviews, trusted count), filtered active packages
// with package-level ratings and seat availability, and the agent's reviews.
func (s *AgentDetailServices) CustomerAgentDetail(req bxapiReq.CustomerAgentDetailReq) (*bxapiResp.CustomerAgentDetailResult, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	agentID := req.AgentID

	// 1) Load the agent profile.
	var agent dxapi.DxAgentProfile
	if err := global.GVA_DB.First(&agent, agentID).Error; err != nil {
		return nil, errors.New("agent not found")
	}
	if agent.Status != constants.AgentStatusActive {
		return nil, fmt.Errorf("agent is not active (status: %s)", agent.Status)
	}

	// 2) Agent-level rating and review count.
	agentRating, agentReviewCount := s.agentRating(agentID)

	// 3) Top destination among active packages.
	topDest := s.topDestinationForAgent(agentID)

	// 4) Paginated active packages, optionally filtered by destination.
	pkgCountDB := global.GVA_DB.Model(&catalog.Package{}).
		Where("agent_id = ? AND status = ?", agentID, constants.PackageStatusActive)
	pkgQuery := global.GVA_DB.Model(&catalog.Package{}).
		Where("agent_id = ? AND status = ?", agentID, constants.PackageStatusActive)
	if req.DestinationCountryID != nil && *req.DestinationCountryID > 0 {
		pkgCountDB = pkgCountDB.Where("destination_country_id = ?", *req.DestinationCountryID)
		pkgQuery = pkgQuery.Where("destination_country_id = ?", *req.DestinationCountryID)
	}
	var packages []catalog.Package
	paginateResult, err := utils.PaginateWithCustomQuery(
		pkgCountDB,
		pkgQuery.Order("approved_at desc"),
		req.Page,
		req.PageSize,
		&packages,
	)
	if err != nil {
		return nil, err
	}
	pkgTotal := paginateResult.Total

	// 5) Agent reviews (reviews where agent_id matches; not package-specific).
	var reviews []engage.Review
	if err := global.GVA_DB.Where("agent_id = ? AND review_status = ?", agentID, constants.ReviewStatusActive).
		Order("created_at desc").Limit(50).Find(&reviews).Error; err != nil {
		return nil, err
	}

	// 6) Build response.
	return &bxapiResp.CustomerAgentDetailResult{
		Agent:           s.buildAgentProfile(agent, agentRating, agentReviewCount),
		Packages:        s.buildAgentPackageCards(packages),
		PackagesTotal:   pkgTotal,
		Reviews:         s.buildAgentReviews(reviews),
		ReviewAvgRating: agentRating,
		ReviewTotal:     agentReviewCount,
		TopDestination:  topDest,
	}, nil
}

// ---------------------------------------------------------------------------
// Private helpers
// ---------------------------------------------------------------------------

func (s *AgentDetailServices) buildAgentProfile(agent dxapi.DxAgentProfile, avgRating float64, reviewCount int) bxapiResp.CustomerAgentProfile {
	statusLabel := agentStatusLabel(agent.Status)
	return bxapiResp.CustomerAgentProfile{
		DxAgentProfile: agent,
		AverageRating:  avgRating,
		ReviewCount:    reviewCount,
		TrustedCount:   agent.TrustedTravelersCount,
		StatusLabel:    statusLabel,
	}
}

func (s *AgentDetailServices) buildAgentPackageCards(packages []catalog.Package) []bxapiResp.CustomerAgentPackageCard {
	out := make([]bxapiResp.CustomerAgentPackageCard, 0, len(packages))
	for _, p := range packages {
		rating, count := s.packageRating(p.ID)
		agentName, agentLogo := s.agentNameAndLogo(p.AgentID)
		salePrice := computeSalePrice(p.DisplayPrice, p.DiscountPercent)
		departureDate := ""
		if p.NextDepartureDate != nil {
			departureDate = p.NextDepartureDate.Format("Jan 02, 2006")
		}
		out = append(out, bxapiResp.CustomerAgentPackageCard{
			Package:         p,
			AgentName:       agentName,
			AgentLogo:       agentLogo,
			SalePrice:       salePrice,
			Rating:          rating,
			ReviewCount:     count,
			SeatsLeft:       p.AvailableSeats,
			DepartureDate:   departureDate,
			NextDepartureAt: p.NextDepartureDate,
		})
	}
	return out
}

func (s *AgentDetailServices) buildAgentReviews(reviews []engage.Review) []bxapiResp.CustomerAgentReviewItem {
	out := make([]bxapiResp.CustomerAgentReviewItem, 0, len(reviews))
	for _, r := range reviews {
		out = append(out, bxapiResp.CustomerAgentReviewItem{
			ID:         r.ID,
			UserName:   s.userName(r.UserID),
			Rating:     r.Rating,
			ReviewText: r.ReviewText,
			CreatedAt:  r.CreatedAt,
		})
	}
	return out
}

func (s *AgentDetailServices) agentRating(agentID uint) (float64, int) {
	var avg *float64
	var count int64
	if err := global.GVA_DB.Model(&engage.Review{}).
		Where("agent_id = ? AND review_status = ?", agentID, constants.ReviewStatusActive).
		Select("AVG(rating) as avg, COUNT(*) as cnt").
		Row().Scan(&avg, &count); err != nil {
		return 0, 0
	}
	r := 0.0
	if avg != nil {
		r = *avg
	}
	return round2(r), int(count)
}

func (s *AgentDetailServices) packageRating(packageID uint) (float64, int) {
	var avg *float64
	var count int64
	if err := global.GVA_DB.Model(&engage.Review{}).
		Where("package_id = ? AND review_status = ?", packageID, constants.ReviewStatusActive).
		Select("AVG(rating) as avg, COUNT(*) as cnt").
		Row().Scan(&avg, &count); err != nil {
		return 0, 0
	}
	r := 0.0
	if avg != nil {
		r = *avg
	}
	return round2(r), int(count)
}

func (s *AgentDetailServices) agentNameAndLogo(agentID uint) (name, logo string) {
	var a dxapi.DxAgentProfile
	if err := global.GVA_DB.Select("company_name, header_img").
		Where("id = ?", agentID).First(&a).Error; err != nil {
		return "", ""
	}
	return a.CompanyName, a.HeaderImg
}

// topDestinationForAgent returns the most common destination country name among
// the agent's active packages, or "" if none.
func (s *AgentDetailServices) topDestinationForAgent(agentID uint) string {
	type destRow struct {
		CountryName string `gorm:"column:country_name"`
		Cnt         int    `gorm:"column:cnt"`
	}
	var row destRow
	if err := global.GVA_DB.Model(&catalog.Package{}).
		Joins("JOIN base_country ON base_country.id = catalog_package.destination_country_id").
		Where("catalog_package.agent_id = ? AND catalog_package.status = ?", agentID, constants.PackageStatusActive).
		Select("base_country.name_en AS country_name").
		Group("base_country.id, base_country.name_en").
		Order("COUNT(catalog_package.id) desc").
		Limit(1).
		Scan(&row).Error; err != nil {
		return ""
	}
	return row.CountryName
}

func (s *AgentDetailServices) userName(userID uint) string {
	var name string
	if err := global.GVA_DB.Model(&bxapiModel.BxsUserProfile{}).
		Select("full_name").
		Where("id = ?", userID).
		Scan(&name).Error; err != nil {
		return ""
	}
	return strings.TrimSpace(name)
}

func computeSalePrice(displayPrice, discountPercentBps int64) int64 {
	if discountPercentBps <= 0 {
		return displayPrice
	}
	return displayPrice - (displayPrice * discountPercentBps / 10000)
}

func agentStatusLabel(status string) string {
	switch status {
	case constants.AgentStatusActive:
		return "Active"
	case constants.AgentStatusPending:
		return "Pending Approval"
	case constants.AgentStatusSuspended:
		return "Suspended"
	case constants.AgentStatusRejected:
		return "Rejected"
	default:
		return status
	}
}

func round2(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}
