package bxapi

import (
	"errors"
	"fmt"

	"github.com/mo7amed4522/tayara_api/constants"
	"github.com/mo7amed4522/tayara_api/global"
	bxapiReq "github.com/mo7amed4522/tayara_api/model/bxapi/request"
	"github.com/mo7amed4522/tayara_api/model/bxapi/response"
	"github.com/mo7amed4522/tayara_api/model/catalog"
	"github.com/mo7amed4522/tayara_api/model/dxapi"
	"github.com/mo7amed4522/tayara_api/model/engage"
	"github.com/mo7amed4522/tayara_api/utils"
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

	// agents: active agents whose company/responsible name OR country matches
	var agents []dxapi.DxAgentProfile
	if err := global.GVA_DB.Where(
		"status = ? AND (company_name LIKE ? OR responsible_name LIKE ? OR country LIKE ?)",
		constants.AgentStatusActive, kw, kw, kw,
	).Order("created_at desc").Limit(10).Find(&agents).Error; err != nil {
		return nil, err
	}

	// packages: active packages whose name, category, location OR destination country name matches
	var packages []catalog.Package
	if err := global.GVA_DB.Where(
		"status = ? AND (name LIKE ? OR category LIKE ? OR location LIKE ? OR destination_country_id IN "+
			"(SELECT id FROM base_country WHERE name_en LIKE ? OR name_ar LIKE ?))",
		constants.PackageStatusActive, kw, kw, kw, kw, kw,
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
		Packages:   s.buildPackageHits(packages),
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
func (s *HomeServices) ListDestinations(page, pageSize int) (*utils.PaginateResult, error) {
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

	// Count query (subquery)
	countSQL := `
		SELECT COUNT(*)
		FROM (
			SELECT c.id, c.name_en, c.code, c.flag_url, COUNT(p.id) AS package_count, MIN(p.display_price) AS min_display_price
			FROM base_country c
			INNER JOIN catalog_package p ON p.destination_country_id = c.id AND p.status = 'active'
			GROUP BY c.id, c.name_en, c.code, c.flag_url
		) as sub
	`
	offset := (page - 1) * pageSize
	dataSQL := fmt.Sprintf(`
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
		LIMIT %d OFFSET %d
	`, pageSize, offset)

	countDB := global.GVA_DB.Raw(countSQL)
	dataDB := global.GVA_DB.Raw(dataSQL)

	// Paginate
	paginateResult, err := utils.PaginateWithCustomQuery(countDB, dataDB, page, pageSize, &rows)
	if err != nil {
		return nil, err
	}

	// Convert raw rows into response cards
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
	paginateResult.Data = out

	return paginateResult, nil
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
		AgentID          uint    `gorm:"column:agent_id"`
		CompanyName      string  `gorm:"column:company_name"`
		HeaderImg        string  `gorm:"column:header_img"`
		Phone            string  `gorm:"column:phone"`
		PhoneCountryCode string  `gorm:"column:phone_country_code"`
		BusinessEmail    string  `gorm:"column:business_email"`
		Country          string  `gorm:"column:country"`
		CompanyAddress   string  `gorm:"column:company_address"`
		LicenseNumber    string  `gorm:"column:license_number"`
		Status           string  `gorm:"column:status"`
		AvgRating        float64 `gorm:"column:avg_rating"`
		ReviewCount      int     `gorm:"column:review_count"`
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

	// 1) wishlisted packages (ordered by created_at desc)
	var wishItems []engage.Wishlist
	if err := global.GVA_DB.Where("user_id = ?", userID).
		Order("created_at desc").Limit(limit).Find(&wishItems).Error; err != nil {
		return nil, err
	}

	// Collect unique package IDs from wishlist, preserving order
	seenPkg := make(map[uint]bool)
	var pkgIDs []uint
	for _, w := range wishItems {
		if !seenPkg[w.PackageID] {
			seenPkg[w.PackageID] = true
			pkgIDs = append(pkgIDs, w.PackageID)
		}
	}

	// 2) Batch load wishlisted packages (only active ones)
	var wishPkgs []catalog.Package
	if err := global.GVA_DB.Where("id IN ? AND status = ?", pkgIDs, constants.PackageStatusActive).
		Find(&wishPkgs).Error; err != nil {
		return nil, err
	}
	pkgByID := make(map[uint]catalog.Package, len(wishPkgs))
	for _, p := range wishPkgs {
		pkgByID[p.ID] = p
	}

	// 3) Batch load agent info and ratings for wishlisted packages
	var wishAgentIDs []uint
	wishAgentSet := make(map[uint]bool)
	for _, p := range wishPkgs {
		if !wishAgentSet[p.AgentID] {
			wishAgentSet[p.AgentID] = true
			wishAgentIDs = append(wishAgentIDs, p.AgentID)
		}
	}
	wishAgentInfo := s.batchAgentNamesAndLogos(wishAgentIDs)
	wishPkgRatings, wishPkgCounts := s.batchPackageRatings(pkgIDs)

	// 4) Build wishlisted results (preserving wish order)
	seen := make(map[uint]bool)
	out := make([]response.CustomerRecommendedResult, 0, limit)
	for _, w := range wishItems {
		pkg, ok := pkgByID[w.PackageID]
		if !ok {
			continue // package not found or not active
		}
		seen[pkg.ID] = true
		info, _ := wishAgentInfo[pkg.AgentID]
		out = append(out, response.CustomerRecommendedResult{
			Package:     pkg,
			AgentName:   info.name,
			AgentLogo:   info.logo,
			Rating:      wishPkgRatings[pkg.ID],
			ReviewCount: wishPkgCounts[pkg.ID],
			Wishlisted:  true,
		})
	}

	// 5) Fill remainder with top approved packages not in wishlist
	//    Reuses pkgIDs from step 1 — no extra wishlist query needed.
	if len(out) < limit {
		var extra []catalog.Package
		if err := global.GVA_DB.Where("status = ? AND id NOT IN ?", constants.PackageStatusActive, pkgIDs).
			Order("approved_at desc").Limit(limit - len(out)).Find(&extra).Error; err != nil {
			return nil, err
		}
		// Batch load agent info and ratings for extras
		var extraAgentIDs []uint
		extraAgentSet := make(map[uint]bool)
		for _, p := range extra {
			if !extraAgentSet[p.AgentID] {
				extraAgentSet[p.AgentID] = true
				extraAgentIDs = append(extraAgentIDs, p.AgentID)
			}
		}
		extraAgentInfo := s.batchAgentNamesAndLogos(extraAgentIDs)
		var extraPkgIDs []uint
		for _, p := range extra {
			extraPkgIDs = append(extraPkgIDs, p.ID)
		}
		extraPkgRatings, extraPkgCounts := s.batchPackageRatings(extraPkgIDs)

		for _, p := range extra {
			info, _ := extraAgentInfo[p.AgentID]
			out = append(out, response.CustomerRecommendedResult{
				Package:     p,
				AgentName:   info.name,
				AgentLogo:   info.logo,
				Rating:      extraPkgRatings[p.ID],
				ReviewCount: extraPkgCounts[p.ID],
				Wishlisted:  seen[p.ID],
			})
		}
	}

	return out, nil
}

// ---------------------------------------------------------------------------
// Ads (home top banner)
// ---------------------------------------------------------------------------

// ActiveHomeAds returns active ad creatives for the home top banner placement with pagination.
func (s *HomeServices) ActiveHomeAds(page, pageSize int) ([]response.AdBannerResult, int64, error) {
	if global.GVA_DB == nil {
		return nil, 0, errors.New("db not init")
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	var ads []engage.AdCreative
	var total int64

	countDB := global.GVA_DB.Model(&engage.AdCreative{}).Where("placement = ? AND status = ?",
		constants.AdPlacementHomeTopBanner, constants.AdStatusActive)
	queryDB := global.GVA_DB.Model(&engage.AdCreative{}).Where("placement = ? AND status = ?",
		constants.AdPlacementHomeTopBanner, constants.AdStatusActive).
		Order("display_order asc")

	paginateResult, err := utils.PaginateWithCustomQuery(countDB, queryDB, page, pageSize, &ads)
	if err != nil {
		return nil, 0, err
	}
	total = paginateResult.Total

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
	return out, total, nil
}

// ---------------------------------------------------------------------------
// Private helpers — batch loading
// ---------------------------------------------------------------------------

// agentInfo holds the name and logo loaded for a single agent.
type agentInfo struct {
	name string
	logo string
}

// batchAgentRatings loads avg rating and review count for multiple agents in one query.
func (s *HomeServices) batchAgentRatings(agentIDs []uint) (ratings map[uint]float64, counts map[uint]int) {
	ratings = make(map[uint]float64)
	counts = make(map[uint]int)
	if len(agentIDs) == 0 {
		return
	}
	type row struct {
		AgentID uint    `gorm:"column:agent_id"`
		Avg     float64 `gorm:"column:avg"`
		Cnt     int64   `gorm:"column:cnt"`
	}
	var rows []row
	if err := global.GVA_DB.Model(&engage.Review{}).
		Where("agent_id IN ?", agentIDs).
		Where("review_status = ?", constants.ReviewStatusActive).
		Select("agent_id, ROUND(AVG(rating), 2) AS avg, COUNT(*) AS cnt").
		Group("agent_id").
		Scan(&rows).Error; err != nil {
		return
	}
	for _, r := range rows {
		ratings[r.AgentID] = r.Avg
		counts[r.AgentID] = int(r.Cnt)
	}
	return
}

// batchPackageRatings loads avg rating and review count for multiple packages in one query.
func (s *HomeServices) batchPackageRatings(pkgIDs []uint) (ratings map[uint]float64, counts map[uint]int) {
	ratings = make(map[uint]float64)
	counts = make(map[uint]int)
	if len(pkgIDs) == 0 {
		return
	}
	type row struct {
		PackageID uint    `gorm:"column:package_id"`
		Avg       float64 `gorm:"column:avg"`
		Cnt       int64   `gorm:"column:cnt"`
	}
	var rows []row
	if err := global.GVA_DB.Model(&engage.Review{}).
		Where("package_id IN ?", pkgIDs).
		Where("review_status = ?", constants.ReviewStatusActive).
		Select("package_id, ROUND(AVG(rating), 2) AS avg, COUNT(*) AS cnt").
		Group("package_id").
		Scan(&rows).Error; err != nil {
		return
	}
	for _, r := range rows {
		ratings[r.PackageID] = r.Avg
		counts[r.PackageID] = int(r.Cnt)
	}
	return
}

// batchAgentNamesAndLogos loads company name and header image for multiple agents in one query.
func (s *HomeServices) batchAgentNamesAndLogos(agentIDs []uint) map[uint]agentInfo {
	out := make(map[uint]agentInfo, len(agentIDs))
	if len(agentIDs) == 0 {
		return out
	}
	type row struct {
		ID          uint   `gorm:"column:id"`
		CompanyName string `gorm:"column:company_name"`
		HeaderImg   string `gorm:"column:header_img"`
	}
	var rows []row
	if err := global.GVA_DB.Model(&dxapi.DxAgentProfile{}).
		Where("id IN ?", agentIDs).
		Select("id, company_name, header_img").
		Scan(&rows).Error; err != nil {
		return out
	}
	for _, r := range rows {
		out[r.ID] = agentInfo{name: r.CompanyName, logo: r.HeaderImg}
	}
	return out
}

// ---------------------------------------------------------------------------
// Private helpers — result builders
// ---------------------------------------------------------------------------

func (s *HomeServices) buildAgentHits(agents []dxapi.DxAgentProfile) []response.CustomerAgentResult {
	var agentIDs []uint
	for _, a := range agents {
		agentIDs = append(agentIDs, a.ID)
	}
	ratings, counts := s.batchAgentRatings(agentIDs)

	out := make([]response.CustomerAgentResult, 0, len(agents))
	for _, a := range agents {
		out = append(out, response.CustomerAgentResult{
			ID:          a.ID,
			CompanyName: a.CompanyName,
			HeaderImg:   a.HeaderImg,
			Rating:      ratings[a.ID],
			ReviewCount: counts[a.ID],
			Status:      a.Status,
		})
	}
	return out
}

func (s *HomeServices) buildPackageHits(packages []catalog.Package) []response.CustomerPackageResult {
	var agentIDs []uint
	var pkgIDs []uint
	agentSet := make(map[uint]bool)
	for _, p := range packages {
		pkgIDs = append(pkgIDs, p.ID)
		if !agentSet[p.AgentID] {
			agentSet[p.AgentID] = true
			agentIDs = append(agentIDs, p.AgentID)
		}
	}
	agentInfo := s.batchAgentNamesAndLogos(agentIDs)
	pkgRatings, pkgCounts := s.batchPackageRatings(pkgIDs)

	out := make([]response.CustomerPackageResult, 0, len(packages))
	for _, p := range packages {
		info, _ := agentInfo[p.AgentID]
		out = append(out, response.CustomerPackageResult{
			ID:              p.ID,
			PackageCode:     p.PackageCode,
			AgentID:         p.AgentID,
			CompanyName:     info.name,
			Name:            p.Name,
			CoverImageURL:   p.CoverImageURL,
			Category:        p.Category,
			DaysDuration:    p.DaysDuration,
			DisplayPrice:    p.DisplayPrice,
			DiscountPercent: p.DiscountPercent,
			Rating:          pkgRatings[p.ID],
			ReviewCount:     pkgCounts[p.ID],
			Status:          p.Status,
		})
	}
	return out
}

func (s *HomeServices) buildExploreHits(packages []catalog.Package) []response.CustomerExploreResult {
	var agentIDs []uint
	var pkgIDs []uint
	agentSet := make(map[uint]bool)
	for _, p := range packages {
		pkgIDs = append(pkgIDs, p.ID)
		if !agentSet[p.AgentID] {
			agentSet[p.AgentID] = true
			agentIDs = append(agentIDs, p.AgentID)
		}
	}
	agentInfo := s.batchAgentNamesAndLogos(agentIDs)
	pkgRatings, pkgCounts := s.batchPackageRatings(pkgIDs)

	out := make([]response.CustomerExploreResult, 0, len(packages))
	for _, p := range packages {
		info, _ := agentInfo[p.AgentID]
		out = append(out, response.CustomerExploreResult{
			ID:              p.ID,
			PackageCode:     p.PackageCode,
			AgentID:         p.AgentID,
			CompanyName:     info.name,
			Name:            p.Name,
			CoverImageURL:   p.CoverImageURL,
			Category:        p.Category,
			DaysDuration:    p.DaysDuration,
			DisplayPrice:    p.DisplayPrice,
			DiscountPercent: p.DiscountPercent,
			Description:     s.truncateDesc(p.Description, 120),
			Rating:          pkgRatings[p.ID],
			ReviewCount:     pkgCounts[p.ID],
			Status:          p.Status,
		})
	}
	return out
}

func (s *HomeServices) buildExploreTripCards(packages []catalog.Package) []response.CustomerExploreTripCard {
	var agentIDs []uint
	var pkgIDs []uint
	agentSet := make(map[uint]bool)
	for _, p := range packages {
		pkgIDs = append(pkgIDs, p.ID)
		if !agentSet[p.AgentID] {
			agentSet[p.AgentID] = true
			agentIDs = append(agentIDs, p.AgentID)
		}
	}
	agentInfo := s.batchAgentNamesAndLogos(agentIDs)
	pkgRatings, pkgCounts := s.batchPackageRatings(pkgIDs)

	out := make([]response.CustomerExploreTripCard, 0, len(packages))
	for _, p := range packages {
		info, _ := agentInfo[p.AgentID]
		out = append(out, response.CustomerExploreTripCard{
			ID:              p.ID,
			PackageCode:     p.PackageCode,
			AgentID:         p.AgentID,
			CompanyName:     info.name,
			Name:            p.Name,
			CoverImageURL:   p.CoverImageURL,
			Category:        p.Category,
			DaysDuration:    p.DaysDuration,
			DisplayPrice:    p.DisplayPrice,
			DiscountPercent: p.DiscountPercent,
			Rating:          pkgRatings[p.ID],
			ReviewCount:     pkgCounts[p.ID],
			Status:          p.Status,
		})
	}
	return out
}

func (s *HomeServices) buildDealResults(packages []catalog.Package) []response.CustomerDealResult {
	var agentIDs []uint
	agentSet := make(map[uint]bool)
	for _, p := range packages {
		if !agentSet[p.AgentID] {
			agentSet[p.AgentID] = true
			agentIDs = append(agentIDs, p.AgentID)
		}
	}
	agentInfo := s.batchAgentNamesAndLogos(agentIDs)

	out := make([]response.CustomerDealResult, 0, len(packages))
	for _, p := range packages {
		info, _ := agentInfo[p.AgentID]
		out = append(out, response.CustomerDealResult{
			Package:           p,
			AgentName:         info.name,
			AgentLogo:         info.logo,
			DiscountPercent:   p.DiscountPercent,
			ViewCount:         0, // view count not tracked yet; field reserved
		})
	}
	return out
}

func (s *HomeServices) truncateDesc(text string, max int) string {
	runes := []rune(text)
	if len(runes) <= max {
		return text
	}
	return string(runes[:max]) + "…"
}
