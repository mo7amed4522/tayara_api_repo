package bxapi

import (
	"errors"
	"fmt"
	"time"

	"github.com/mo7amed4522/tayara_api/constants"
	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/model/base"
	bxapiReq "github.com/mo7amed4522/tayara_api/model/bxapi/request"
	"github.com/mo7amed4522/tayara_api/model/bxapi/response"
	"github.com/mo7amed4522/tayara_api/model/catalog"
	"github.com/mo7amed4522/tayara_api/model/dxapi"
	"github.com/mo7amed4522/tayara_api/model/engage"
	"github.com/mo7amed4522/tayara_api/utils"
	"github.com/mo7amed4522/tayara_api/utils/cache"
)

// HomeServices owns the public/buyer-facing home-screen data aggregations.
type HomeServices struct{}

// ---------------------------------------------------------------------------
// Search
// ---------------------------------------------------------------------------

// CustomerSearch runs a keyword search across active agents, plus top explore
// packages and similar destinations from the same country, returning the shape the mobile home search expects.
func (s *HomeServices) CustomerSearch(req bxapiReq.CustomerSearchReq) (*response.CustomerSearchResult, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	kw := "%" + req.Keyword + "%"

	// ── 1. Resolve keyword → city (preferred) then country (fallback) ─────
	var (
		cityID    uint
		cityName  string
		countryID uint
		foundCity bool
	)

	// Try city first so "Dubai" hits the Dubai city row, not the AE country row.
	if id, _ := s.findCityByKeyword(req.Keyword); id > 0 {
		var c base.City
		if err := global.GVA_DB.Select("id", "name", "country_id").
			First(&c, id).Error; err == nil {
			cityID = c.ID
			cityName = c.Name
			countryID = c.CountryID
			foundCity = true
		}
	}

	// Fallback: city by name (in case keyword JSONMap didn't match).
	if !foundCity {
		var c base.City
		if err := global.GVA_DB.Where("name LIKE ? AND is_active = ?", kw, true).
			Select("id", "name", "country_id").Limit(1).First(&c).Error; err == nil {
			cityID = c.ID
			cityName = c.Name
			countryID = c.CountryID
			foundCity = true
		}
	}

	// If no city, try country.
	if countryID == 0 {
		countryID = s.findCountryByKeyword(req.Keyword)
	}
	if countryID == 0 {
		var c base.Country
		if err := global.GVA_DB.Where("name_en LIKE ? OR name_ar LIKE ? OR code LIKE ?", kw, kw, kw).
			Select("id").Limit(1).First(&c).Error; err == nil {
			countryID = c.ID
		}
	}

	// ── 2. Agents: text-only matching (agent table has no city/country FK) ─
	agentQ := global.GVA_DB.Where("status = ?", constants.AgentStatusActive)
	if foundCity {
		agentQ = agentQ.Where(
			"company_name LIKE ? OR responsible_name LIKE ? OR country LIKE ? OR city LIKE ?",
			kw, kw, kw, "%"+cityName+"%",
		)
	} else {
		agentQ = agentQ.Where(
			"company_name LIKE ? OR responsible_name LIKE ? OR country LIKE ?",
			kw, kw, kw,
		)
	}
	var agents []dxapi.DxAgentProfile
	if err := agentQ.Order("created_at desc").Limit(10).Find(&agents).Error; err != nil {
		return nil, err
	}

	// ── 3. topExplore: packages strictly for the searched city ────────────
	var topExplore []catalog.Package
	topQ := global.GVA_DB.Where("status = ?", constants.PackageStatusActive)
	if foundCity {
		topQ = topQ.Where(
			"city_id = ? OR location LIKE ? OR name LIKE ?",
			cityID, "%"+cityName+"%", "%"+cityName+"%",
		)
	} else if countryID > 0 {
		topQ = topQ.Where("destination_country_id = ?", countryID)
	}
	if err := topQ.Order("approved_at desc").Limit(6).Find(&topExplore).Error; err != nil {
		return nil, err
	}

	// ── 4. similarDestinations: same country EXCLUDING the searched city ──
	var similarDestinations []catalog.Package
	if countryID > 0 {
		simQ := global.GVA_DB.Where(
			"status = ? AND destination_country_id = ?",
			constants.PackageStatusActive, countryID,
		)
		if foundCity {
			simQ = simQ.Where(
				"(city_id IS NULL OR city_id <> ?) AND NOT (location LIKE ? OR name LIKE ?)",
				cityID, "%"+cityName+"%", "%"+cityName+"%",
			)
		}
		if err := simQ.Order("approved_at desc").Limit(10).Find(&similarDestinations).Error; err != nil {
			return nil, err
		}
	}

	return &response.CustomerSearchResult{
		Keyword:             req.Keyword,
		Agents:              s.buildAgentHits(agents),
		TopExplore:          s.buildExploreHits(topExplore),
		SimilarDestinations: s.buildExploreHits(similarDestinations),
	}, nil
}

// findCountryByKeyword searches for a country by keyword in the Keywords JSONMap field
// Returns country ID if found, 0 otherwise
func (s *HomeServices) findCountryByKeyword(keyword string) uint {
	cacheKey := "bxapi:country:keyword:" + keyword
	if val, found, _ := cache.Get(cacheKey); found {
		if id, ok := val.(uint); ok {
			return id
		}
	}

	var countryID uint
	// keywords structure: {"keywords": ["UAE", "United Arab Emirates", "Emirates", "AE"]}
	kw := "\"" + keyword + "\""
	err := global.GVA_DB.Model(&base.Country{}).
		Where("JSON_CONTAINS(keywords, ?, '$.keywords')", kw).
		Select("id").
		Limit(1).
		Scan(&countryID).Error
	if err != nil {
		return 0
	}

	cache.Set(cacheKey, countryID, 8, cache.AddJitter(5*time.Minute))
	return countryID
}

// findCityByKeyword searches for a city by keyword in the Keywords JSONMap field
// Returns city ID and city name if found, 0 and "" otherwise
func (s *HomeServices) findCityByKeyword(keyword string) (uint, string) {
	cacheKey := "bxapi:city:keyword:" + keyword
	if val, found, _ := cache.Get(cacheKey); found {
		if result, ok := val.(struct {
			ID   uint
			Name string
		}); ok {
			return result.ID, result.Name
		}
	}

	type cityResult struct {
		ID   uint
		Name string
	}
	var result cityResult
	kw := "\"" + keyword + "\""
	err := global.GVA_DB.Model(&base.City{}).
		Where("JSON_CONTAINS(keywords, ?, '$.keywords') AND is_active = ?", kw, true).
		Select("id", "name").
		Limit(1).
		Scan(&result).Error
	if err != nil {
		return 0, ""
	}

	cache.Set(cacheKey, result, 64, cache.AddJitter(5*time.Minute))
	return result.ID, result.Name
}

// ---------------------------------------------------------------------------
// Explore trips (cards)
// ---------------------------------------------------------------------------

// ListExploreTrips returns a paginated list of trending active packages.
func (s *HomeServices) ListExploreTrips(req bxapiReq.CustomerHomeExploreReq) ([]response.CustomerExploreTripCard, int64, int64, error) {
	if global.GVA_DB == nil {
		return nil, 0, 0, errors.New("db not init")
	}
	countDB := global.GVA_DB.Model(&catalog.Package{}).
		Where("status = ?", constants.PackageStatusActive)
	if req.Category != "" {
		countDB = countDB.Where("category = ?", req.Category)
	}
	queryDB := global.GVA_DB.Model(&catalog.Package{}).
		Where("status = ?", constants.PackageStatusActive)
	if req.Category != "" {
		queryDB = queryDB.Where("category = ?", req.Category)
	}
	var packages []catalog.Package
	pr, err := utils.PaginateWithCustomQuery(countDB, queryDB.Order("approved_at desc"), req.Page, req.PageSize, &packages)
	if err != nil {
		return nil, 0, 0, err
	}
	return s.buildExploreTripCards(packages), pr.Total, pr.TotalPage, nil
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
		CountryID       uint   `gorm:"column:country_id"`
		CountryName     string `gorm:"column:country_name"`
		CountryCode     string `gorm:"column:country_code"`
		PackageCount    int    `gorm:"column:package_count"`
		MinDisplayPrice int64  `gorm:"column:min_display_price"`
	}
	var rows []destRow

	countDB := global.GVA_DB.Model(&base.Country{}).
		Joins("JOIN catalog_package ON catalog_package.destination_country_id = base_country.id").
		Where("catalog_package.status = ?", constants.PackageStatusActive).
		Distinct("base_country.id")
	queryDB := global.GVA_DB.Model(&base.Country{}).
		Joins("JOIN catalog_package ON catalog_package.destination_country_id = base_country.id").
		Where("catalog_package.status = ?", constants.PackageStatusActive).
		Select("base_country.id AS country_id, base_country.name_en AS country_name, base_country.code AS country_code, COUNT(catalog_package.id) AS package_count, MIN(catalog_package.display_price) AS min_display_price").
		Group("base_country.id, base_country.name_en, base_country.code").
		Order("package_count desc, min_display_price asc")
	paginateResult, err := utils.PaginateWithCustomQuery(countDB, queryDB, page, pageSize, &rows)
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

// ListDeals returns a paginated list of discounted active packages.
func (s *HomeServices) ListDeals(req bxapiReq.CustomerHomeDealsReq) ([]response.CustomerDealResult, int64, int64, error) {
	if global.GVA_DB == nil {
		return nil, 0, 0, errors.New("db not init")
	}
	countDB := global.GVA_DB.Model(&catalog.Package{}).
		Where("status = ? AND discount_percent > ?", constants.PackageStatusActive, 0)
	if req.Category != "" {
		countDB = countDB.Where("category = ?", req.Category)
	}
	queryDB := global.GVA_DB.Model(&catalog.Package{}).
		Where("status = ? AND discount_percent > ?", constants.PackageStatusActive, 0)
	if req.Category != "" {
		queryDB = queryDB.Where("category = ?", req.Category)
	}
	var packages []catalog.Package
	pr, err := utils.PaginateWithCustomQuery(countDB, queryDB.Order("discount_percent desc, updated_at desc"), req.Page, req.PageSize, &packages)
	if err != nil {
		return nil, 0, 0, err
	}
	return s.buildDealResults(packages), pr.Total, pr.TotalPage, nil
}

// ---------------------------------------------------------------------------
// Top rated agencies
// ---------------------------------------------------------------------------

// ListTopRatedAgents returns active agents ranked by average review rating.
func (s *HomeServices) ListTopRatedAgents(page, pageSize int) (*utils.PaginateResult, error) {
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
	countDB := global.GVA_DB.Model(&dxapi.DxAgentProfile{}).
		Joins("JOIN engage_review ON engage_review.agent_id = dx_agent_profile.id").
		Where("dx_agent_profile.status = ? AND engage_review.review_status = ?", constants.AgentStatusActive, constants.ReviewStatusActive).
		Distinct("dx_agent_profile.id")
	queryDB := global.GVA_DB.Model(&dxapi.DxAgentProfile{}).
		Joins("JOIN engage_review ON engage_review.agent_id = dx_agent_profile.id").
		Where("dx_agent_profile.status = ? AND engage_review.review_status = ?", constants.AgentStatusActive, constants.ReviewStatusActive).
		Select("dx_agent_profile.id AS agent_id, dx_agent_profile.company_name, dx_agent_profile.header_img, dx_agent_profile.phone, dx_agent_profile.phone_country_code, dx_agent_profile.business_email, dx_agent_profile.country, dx_agent_profile.company_address, dx_agent_profile.license_number, dx_agent_profile.status, ROUND(AVG(engage_review.rating), 2) AS avg_rating, COUNT(engage_review.id) AS review_count").
		Group("dx_agent_profile.id, dx_agent_profile.company_name, dx_agent_profile.header_img, dx_agent_profile.phone, dx_agent_profile.phone_country_code, dx_agent_profile.business_email, dx_agent_profile.country, dx_agent_profile.company_address, dx_agent_profile.license_number, dx_agent_profile.status").
		Order("avg_rating desc, review_count desc")
	var rows []agentRatingRow
	paginateResult, err := utils.PaginateWithCustomQuery(countDB, queryDB, page, pageSize, &rows)
	if err != nil {
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
	paginateResult.Data = out
	return paginateResult, nil
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
// Uses cache for individual package ratings, only queries DB for missing entries.
func (s *HomeServices) batchPackageRatings(pkgIDs []uint) (ratings map[uint]float64, counts map[uint]int) {
	ratings = make(map[uint]float64)
	counts = make(map[uint]int)
	if len(pkgIDs) == 0 {
		return
	}

	// Check cache for each package
	var missingIDs []uint
	for _, pkgID := range pkgIDs {
		cacheKey := fmt.Sprintf("bxapi:pkg:rating:%d", pkgID)
		if val, found, _ := cache.Get(cacheKey); found {
			if cached, ok := val.(struct {
				Avg float64
				Cnt int
			}); ok {
				ratings[pkgID] = cached.Avg
				counts[pkgID] = cached.Cnt
				continue
			}
		}
		missingIDs = append(missingIDs, pkgID)
	}

	// Query DB only for missing IDs
	if len(missingIDs) > 0 {
		type row struct {
			PackageID uint    `gorm:"column:package_id"`
			Avg       float64 `gorm:"column:avg"`
			Cnt       int64   `gorm:"column:cnt"`
		}
		var rows []row
		if err := global.GVA_DB.Model(&engage.Review{}).
			Where("package_id IN ?", missingIDs).
			Where("review_status = ?", constants.ReviewStatusActive).
			Select("package_id, ROUND(AVG(rating), 2) AS avg, COUNT(*) AS cnt").
			Group("package_id").
			Scan(&rows).Error; err == nil {
			for _, r := range rows {
				ratings[r.PackageID] = r.Avg
				counts[r.PackageID] = int(r.Cnt)
				// Cache the result
				cacheKey := fmt.Sprintf("bxapi:pkg:rating:%d", r.PackageID)
				cache.Set(cacheKey, struct {
					Avg float64
					Cnt int
				}{Avg: r.Avg, Cnt: int(r.Cnt)}, 24, cache.AddJitter(10*time.Minute))
			}
		}
	}

	return
}

// batchAgentNamesAndLogos loads company name and header image for multiple agents in one query.
// Uses cache for individual agent info, only queries DB for missing entries.
func (s *HomeServices) batchAgentNamesAndLogos(agentIDs []uint) map[uint]agentInfo {
	out := make(map[uint]agentInfo, len(agentIDs))
	if len(agentIDs) == 0 {
		return out
	}

	// Check cache for each agent
	var missingIDs []uint
	for _, agentID := range agentIDs {
		cacheKey := fmt.Sprintf("bxapi:agent:info:%d", agentID)
		if val, found, _ := cache.Get(cacheKey); found {
			if info, ok := val.(agentInfo); ok {
				out[agentID] = info
				continue
			}
		}
		missingIDs = append(missingIDs, agentID)
	}

	// Query DB only for missing IDs
	if len(missingIDs) > 0 {
		type row struct {
			ID          uint   `gorm:"column:id"`
			CompanyName string `gorm:"column:company_name"`
			HeaderImg   string `gorm:"column:header_img"`
		}
		var rows []row
		if err := global.GVA_DB.Model(&dxapi.DxAgentProfile{}).
			Where("id IN ?", missingIDs).
			Select("id, company_name, header_img").
			Scan(&rows).Error; err == nil {
			for _, r := range rows {
				info := agentInfo{name: r.CompanyName, logo: r.HeaderImg}
				out[r.ID] = info
				// Cache the result
				cacheKey := fmt.Sprintf("bxapi:agent:info:%d", r.ID)
				cache.Set(cacheKey, info, 64, cache.AddJitter(10*time.Minute))
			}
		}
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
			Package:         p,
			AgentName:       info.name,
			AgentLogo:       info.logo,
			DiscountPercent: p.DiscountPercent,
			ViewCount:       0, // view count not tracked yet; field reserved
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
