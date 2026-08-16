package bxapi

import (
	"errors"

	"github.com/mo7amed4522/tayara_api/constants"
	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/model/bxapi/response"
	"github.com/mo7amed4522/tayara_api/model/catalog"
	"github.com/mo7amed4522/tayara_api/model/dxapi"
	"github.com/mo7amed4522/tayara_api/model/engage"
	bxapiReq "github.com/mo7amed4522/tayara_api/model/bxapi/request"
)

// HomeServices owns the public/buyer-facing home-screen data aggregations.
type HomeServices struct{}

// ---------------------------------------------------------------------------
// Search
// ---------------------------------------------------------------------------

// CustomerSearch runs a keyword search across active agents and packages, plus top explore
// packages, returning the shape the mobile home search expects.
func (s *HomeServices) CustomerSearch(req bxapiReq.CustomerSearchReq) (*response.CustomerSearchResult, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	kw := "%" + req.Keyword + "%"

	// agents: active agents whose company name matches
	var agents []dxapi.DxAgentProfile
	if err := global.GVA_DB.Where(
		"status = ? AND (company_name LIKE ? OR responsible_name LIKE ?)",
		constants.AgentStatusActive, kw, kw,
	).Order("created_at desc").Limit(10).Find(&agents).Error; err != nil {
		return nil, err
	}

	// packages: active packages whose name or category matches
	var packages []catalog.Package
	if err := global.GVA_DB.Where(
		"status = ? AND (name LIKE ? OR category LIKE ?)",
		constants.PackageStatusActive, kw, kw,
	).Order("updated_at desc").Limit(20).Find(&packages).Error; err != nil {
		return nil, err
	}

	// top explore: most-recently approved active packages (up to 6)
	var topExplore []catalog.Package
	if err := global.GVA_DB.Where("status = ?", constants.PackageStatusActive).
		Order("approved_at desc").Limit(6).Find(&topExplore).Error; err != nil {
		return nil, err
	}

	return &response.CustomerSearchResult{
		Keyword:    req.Keyword,
		Agents:     s.buildAgentHits(agents),
		Packages:   s.buildPackageHits(packages, false),
		TopExplore: s.buildExploreHits(topExplore),
	}, nil
}

// ---------------------------------------------------------------------------
// Explore trips (cards)
// ---------------------------------------------------------------------------

// ListExploreTrips returns trending active packages (with agent info + aggregated rating).
func (s *HomeServices) ListExploreTrips(req bxapiReq.CustomerHomeExploreReq) ([]response.CustomerExploreTripCard, int64, error) {
	if global.GVA_DB == nil {
		return nil, 0, errors.New("db not init")
	}
	db := global.GVA_DB.Model(&catalog.Package{}).
		Where("status = ?", constants.PackageStatusActive)
	if req.Category != "" {
		db = db.Where("category = ?", req.Category)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var packages []catalog.Package
	if err := db.Scopes(req.Paginate()).Order("approved_at desc").Find(&packages).Error; err != nil {
		return nil, 0, err
	}
	return s.buildExploreTripCards(packages), total, nil
}

// ---------------------------------------------------------------------------
// Discover destinations
// ---------------------------------------------------------------------------

// ListDestinations returns countries that have at least one active package, with package count
// and the lowest display price available.
func (s *HomeServices) ListDestinations() ([]response.CustomerDestinationCard, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	type destRow struct {
		CountryID       uint  `gorm:"column:country_id"`
		CountryName     string `gorm:"column:country_name"`
		CountryCode     string `gorm:"column:country_code"`
		FlagURL         string `gorm:"column:flag_url"`
		PackageCount    int   `gorm:"column:package_count"`
		MinDisplayPrice int64 `gorm:"column:min_display_price"`
	}
	var rows []destRow
	query := `
		SELECT
			c.id AS country_id,
			c.name_en AS country_name,
			c.code AS country_code,
			c.flag_url AS flag_url,
			COUNT(p.id) AS package_count,
			MIN(p.display_price) AS min_display_price
		FROM base_country c
		INNER JOIN catalog_package p ON p.destination_country_id = c.id AND p.status = 'active'
		GROUP BY c.id, c.name_en, c.code, c.flag_url
		ORDER BY package_count DESC, min_display_price ASC
		LIMIT 20
	`
	if err := global.GVA_DB.Raw(query).Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]response.CustomerDestinationCard, 0, len(rows))
	for _, r := range rows {
		out = append(out, response.CustomerDestinationCard{
			CountryID:       r.CountryID,
			CountryName:     r.CountryName,
			CountryCode:     r.CountryCode,
			FlagURL:         r.FlagURL,
			PackageCount:    r.PackageCount,
			MinDisplayPrice: r.MinDisplayPrice,
		})
	}
	return out, nil
}

// ---------------------------------------------------------------------------
// Deals
// ---------------------------------------------------------------------------

// ListDeals returns active packages that have a discount applied, ordered by discount desc.
func (s *HomeServices) ListDeals(req bxapiReq.CustomerHomeDealsReq) ([]response.CustomerDealResult, int64, error) {
	if global.GVA_DB == nil {
		return nil, 0, errors.New("db not init")
	}
	db := global.GVA_DB.Model(&catalog.Package{}).
		Where("status = ? AND discount_percent > ?", constants.PackageStatusActive, 0)
	if req.Category != "" {
		db = db.Where("category = ?", req.Category)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var packages []catalog.Package
	if err := db.Scopes(req.Paginate()).Order("discount_percent desc, updated_at desc").Find(&packages).Error; err != nil {
		return nil, 0, err
	}
	return s.buildDealResults(packages), total, nil
}

// ---------------------------------------------------------------------------
// Top rated agencies
// ---------------------------------------------------------------------------

// ListTopRatedAgents returns active agents ranked by average review rating.
func (s *HomeServices) ListTopRatedAgents(limit int) ([]response.CustomerAgentCard, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	type agentRatingRow struct {
		AgentID       uint     `gorm:"column:agent_id"`
		CompanyName   string   `gorm:"column:company_name"`
		HeaderImg     string   `gorm:"column:header_img"`
		Phone         string   `gorm:"column:phone"`
		PhoneCountryCode string `gorm:"column:phone_country_code"`
		BusinessEmail string   `gorm:"column:business_email"`
		Country       string   `gorm:"column:country"`
		CompanyAddress string  `gorm:"column:company_address"`
		LicenseNumber string   `gorm:"column:license_number"`
		Status        string   `gorm:"column:status"`
		AvgRating     float64  `gorm:"column:avg_rating"`
		ReviewCount   int      `gorm:"column:review_count"`
	}
	query := `
		SELECT
			a.id AS agent_id,
			a.company_name,
			a.header_img,
			a.phone,
			a.phone_country_code,
			a.business_email,
			a.country,
			a.company_address,
			a.license_number,
			a.status,
			ROUND(AVG(r.rating), 2) AS avg_rating,
			COUNT(r.id) AS review_count
		FROM dx_agent_profile a
		INNER JOIN engage_review r ON r.agent_id = a.id AND r.review_status = 'active'
		WHERE a.status = 'active'
		GROUP BY a.id, a.company_name, a.header_img, a.phone, a.phone_country_code,
		         a.business_email, a.country, a.company_address, a.license_number, a.status
		ORDER BY avg_rating DESC, review_count DESC
		LIMIT ?
	`
	var rows []agentRatingRow
	if err := global.GVA_DB.Raw(query, limit).Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]response.CustomerAgentCard, 0, len(rows))
	for _, r := range rows {
		agent := dxapi.DxAgentProfile{
			ID: r.AgentID, CompanyName: r.CompanyName, HeaderImg: r.HeaderImg,
			Phone: r.Phone, PhoneCountryCode: r.PhoneCountryCode, BusinessEmail: r.BusinessEmail,
			Country: r.Country, CompanyAddress: r.CompanyAddress, LicenseNumber: r.LicenseNumber,
			Status: r.Status,
		}
		out = append(out, response.CustomerAgentCard{
			DxAgentProfile: agent,
			AverageRating:  r.AvgRating,
			ReviewCount:    r.ReviewCount,
		})
	}
	return out, nil
}

// ---------------------------------------------------------------------------
// Recommended for you (authenticated)
// ---------------------------------------------------------------------------

// RecommendedForYou returns packages the user has wishlisted, plus a small pool of
// recommended active packages ordered by approval/recency.
func (s *HomeServices) RecommendedForYou(userID uint, limit int) ([]response.CustomerRecommendedResult, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}

	// 1) wishlisted packages
	var wishItems []engage.Wishlist
	if err := global.GVA_DB.Where("user_id = ?", userID).
		Order("created_at desc").Limit(limit).Find(&wishItems).Error; err != nil {
		return nil, err
	}

	seen := make(map[uint]bool)
	out := make([]response.CustomerRecommendedResult, 0, limit)

	for _, w := range wishItems {
		var pkg catalog.Package
		if err := global.GVA_DB.Where("id = ? AND status = ?", w.PackageID, constants.PackageStatusActive).
			First(&pkg).Error; err != nil {
			continue
		}
		seen[pkg.ID] = true
		res := response.CustomerRecommendedResult{
			Package:    pkg,
			Wishlisted: true,
		}
		res.AgentName, res.AgentLogo = s.agentNameAndLogo(pkg.AgentID)
		s.fillRatingInto(&res)
		out = append(out, res)
	}

	// 2) fill remainder with top approved packages the user hasn't wishlisted
	if len(out) < limit {
		var extra []catalog.Package
		if err := global.GVA_DB.Where("status = ? AND id NOT IN ?", constants.PackageStatusActive,
			s.wishlistPackageIDs(userID)).
			Order("approved_at desc").Limit(limit - len(out)).Find(&extra).Error; err != nil {
			return nil, err
		}
		for _, pkg := range extra {
			res := response.CustomerRecommendedResult{
				Package:    pkg,
				Wishlisted: seen[pkg.ID],
			}
			res.AgentName, res.AgentLogo = s.agentNameAndLogo(pkg.AgentID)
			s.fillRatingInto(&res)
			out = append(out, res)
		}
	}

	return out, nil
}

// ---------------------------------------------------------------------------
// Ads (home top banner)
// ---------------------------------------------------------------------------

// ActiveHomeAds returns active ad creatives for the home top banner placement.
func (s *HomeServices) ActiveHomeAds() ([]response.AdBannerResult, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	var ads []engage.AdCreative
	if err := global.GVA_DB.Where("placement = ? AND status = ?",
		constants.AdPlacementHomeTopBanner, constants.AdStatusActive).
		Order("display_order asc").Find(&ads).Error; err != nil {
		return nil, err
	}
	out := make([]response.AdBannerResult, 0, len(ads))
	for _, ad := range ads {
		out = append(out, response.AdBannerResult{
			ID:              ad.ID,
			AdCode:          ad.AdCode,
			BannerImageURL:  ad.BannerImageURL,
			OverlayTitle:    ad.OverlayTitle,
			OverlaySubtitle: ad.OverlaySubtitle,
			CtaLabel:        ad.CtaLabel,
			RedirectType:    ad.RedirectType,
			TargetCountryID: ad.TargetCountryID,
			TargetAgentID:   ad.TargetAgentID,
			TargetPackageID: ad.TargetPackageID,
			CreatedAt:       ad.CreatedAt,
		})
	}
	return out, nil
}

// ---------------------------------------------------------------------------
// Private helpers
// ---------------------------------------------------------------------------

func (s *HomeServices) buildAgentHits(agents []dxapi.DxAgentProfile) []response.CustomerAgentResult {
	out := make([]response.CustomerAgentResult, 0, len(agents))
	for _, a := range agents {
		rating, count := s.agentRating(a.ID)
		out = append(out, response.CustomerAgentResult{
			ID:          a.ID,
			CompanyName: a.CompanyName,
			HeaderImg:   a.HeaderImg,
			Rating:      rating,
			ReviewCount: count,
			Status:      a.Status,
		})
	}
	return out
}

func (s *HomeServices) buildPackageHits(packages []catalog.Package, withDesc bool) []response.CustomerPackageResult {
	out := make([]response.CustomerPackageResult, 0, len(packages))
	for _, p := range packages {
		rating, count := s.packageRating(p.ID)
		agentName, _ := s.agentNameAndLogo(p.AgentID)
		out = append(out, response.CustomerPackageResult{
			ID:              p.ID,
			PackageCode:     p.PackageCode,
			AgentID:         p.AgentID,
			CompanyName:     agentName,
			Name:            p.Name,
			CoverImageURL:   p.CoverImageURL,
			Category:        p.Category,
			DaysDuration:    p.DaysDuration,
			DisplayPrice:    p.DisplayPrice,
			DiscountPercent: p.DiscountPercent,
			Rating:          rating,
			ReviewCount:     count,
			Status:          p.Status,
		})
	}
	return out
}

func (s *HomeServices) buildExploreHits(packages []catalog.Package) []response.CustomerExploreResult {
	out := make([]response.CustomerExploreResult, 0, len(packages))
	for _, p := range packages {
		rating, count := s.packageRating(p.ID)
		agentName, _ := s.agentNameAndLogo(p.AgentID)
		out = append(out, response.CustomerExploreResult{
			ID:              p.ID,
			PackageCode:     p.PackageCode,
			AgentID:         p.AgentID,
			CompanyName:     agentName,
			Name:            p.Name,
			CoverImageURL:   p.CoverImageURL,
			Category:        p.Category,
			DaysDuration:    p.DaysDuration,
			DisplayPrice:    p.DisplayPrice,
			DiscountPercent: p.DiscountPercent,
			Description:     s.truncateDesc(p.Description, 120),
			Rating:          rating,
			ReviewCount:     count,
			Status:          p.Status,
		})
	}
	return out
}

func (s *HomeServices) buildExploreTripCards(packages []catalog.Package) []response.CustomerExploreTripCard {
	out := make([]response.CustomerExploreTripCard, 0, len(packages))
	for _, p := range packages {
		rating, count := s.packageRating(p.ID)
		agentName, _ := s.agentNameAndLogo(p.AgentID)
		out = append(out, response.CustomerExploreTripCard{
			ID:              p.ID,
			PackageCode:     p.PackageCode,
			AgentID:         p.AgentID,
			CompanyName:     agentName,
			Name:            p.Name,
			CoverImageURL:   p.CoverImageURL,
			Category:        p.Category,
			DaysDuration:    p.DaysDuration,
			DisplayPrice:    p.DisplayPrice,
			DiscountPercent: p.DiscountPercent,
			Rating:          rating,
			ReviewCount:     count,
			Status:          p.Status,
		})
	}
	return out
}

func (s *HomeServices) buildDealResults(packages []catalog.Package) []response.CustomerDealResult {
	out := make([]response.CustomerDealResult, 0, len(packages))
	for _, p := range packages {
		agentName, agentLogo := s.agentNameAndLogo(p.AgentID)
		out = append(out, response.CustomerDealResult{
			Package:           p,
			AgentName:         agentName,
			AgentLogo:         agentLogo,
			DiscountPercent:   p.DiscountPercent,
			ViewCount:         0, // view count not tracked yet; field reserved
		})
	}
	return out
}

func (s *HomeServices) fillRatingInto(res *response.CustomerRecommendedResult) {
	rating, count := s.packageRating(res.ID)
	res.Rating = rating
	res.ReviewCount = count
}

func (s *HomeServices) agentRating(agentID uint) (float64, int) {
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
	return r, int(count)
}

func (s *HomeServices) packageRating(packageID uint) (float64, int) {
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
	return r, int(count)
}

func (s *HomeServices) agentNameAndLogo(agentID uint) (name, logo string) {
	var a dxapi.DxAgentProfile
	if err := global.GVA_DB.Select("company_name, header_img").
		Where("id = ?", agentID).First(&a).Error; err != nil {
		return "", ""
	}
	return a.CompanyName, a.HeaderImg
}

func (s *HomeServices) wishlistPackageIDs(userID uint) []uint {
	var ids []uint
	_ = global.GVA_DB.Model(&engage.Wishlist{}).
		Where("user_id = ?", userID).
		Pluck("package_id", &ids).Error
	return ids
}

func (s *HomeServices) truncateDesc(text string, max int) string {
	runes := []rune(text)
	if len(runes) <= max {
		return text
	}
	return string(runes[:max]) + "…"
}
