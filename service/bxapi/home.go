package bxapi

import (
	"context"
	"errors"

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
	"gorm.io/gorm"
)

// HomeServices owns the public/buyer-facing home-screen data aggregations.
type HomeServices struct{}

type cityKeywordResult struct {
	ID   uint
	Name string
}

// ---------------------------------------------------------------------------
// Search
// ---------------------------------------------------------------------------

// CustomerSearch runs a keyword search across active agents, plus top explore
// packages and similar destinations from the same country, returning the shape the mobile home search expects.
func (s *HomeServices) CustomerSearch(req bxapiReq.CustomerSearchReq) (*response.CustomerSearchResult, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	result, _, err := cache.GetOrLoadJSON(context.Background(), HomeSearchKey(req), sharedResponseTTL, func() (*response.CustomerSearchResult, error) {
		return s.customerSearchUncached(req)
	})
	return result, err
}

func (s *HomeServices) customerSearchUncached(req bxapiReq.CustomerSearchReq) (*response.CustomerSearchResult, error) {
	kw := "%" + req.Keyword + "%"

	// ── 1. Resolve keyword → city (preferred) then country (fallback) ─────
	var (
		cityID    uint
		cityName  string
		countryID uint
		foundCity bool
	)

	// Try city first so "Dubai" hits the Dubai city row, not the AE country row.
	if id, name, err := s.findCityByKeyword(req.Keyword); err != nil {
		return nil, err
	} else if id > 0 {
		var c base.City
		if err := global.GVA_DB.Select("id", "name", "country_id").
			First(&c, id).Error; err == nil {
			cityID = c.ID
			cityName = name
			countryID = c.CountryID
			foundCity = true
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
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
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}

	// If no city, try country.
	if countryID == 0 {
		var err error
		countryID, err = s.findCountryByKeyword(req.Keyword)
		if err != nil {
			return nil, err
		}
	}
	if countryID == 0 {
		var c base.Country
		if err := global.GVA_DB.Where("name_en LIKE ? OR name_ar LIKE ? OR code LIKE ?", kw, kw, kw).
			Select("id").Limit(1).First(&c).Error; err == nil {
			countryID = c.ID
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
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

	agentHits, err := s.buildAgentHits(agents)
	if err != nil {
		return nil, err
	}
	topHits, err := s.buildExploreHits(topExplore)
	if err != nil {
		return nil, err
	}
	similarHits, err := s.buildExploreHits(similarDestinations)
	if err != nil {
		return nil, err
	}
	return &response.CustomerSearchResult{
		Keyword:             req.Keyword,
		Agents:              agentHits,
		TopExplore:          topHits,
		SimilarDestinations: similarHits,
	}, nil
}

// findCountryByKeyword searches for a country by keyword in the Keywords JSONMap field
// Returns country ID if found, 0 otherwise
func (s *HomeServices) findCountryByKeyword(keyword string) (uint, error) {
	var countryID uint
	// keywords structure: {"keywords": ["UAE", "United Arab Emirates", "Emirates", "AE"]}
	kw := "\"" + keyword + "\""
	err := global.GVA_DB.Model(&base.Country{}).
		Where("JSON_CONTAINS(keywords, ?, '$.keywords')", kw).
		Select("id").
		Limit(1).
		Scan(&countryID).Error
	if err != nil {
		return 0, err
	}

	return countryID, nil
}

// findCityByKeyword searches for a city by keyword in the Keywords JSONMap field
// Returns city ID and city name if found, 0 and "" otherwise
func (s *HomeServices) findCityByKeyword(keyword string) (uint, string, error) {
	var result cityKeywordResult
	kw := "\"" + keyword + "\""
	err := global.GVA_DB.Model(&base.City{}).
		Where("JSON_CONTAINS(keywords, ?, '$.keywords') AND is_active = ?", kw, true).
		Select("id", "name").
		Limit(1).
		Scan(&result).Error
	if err != nil {
		return 0, "", err
	}

	return result.ID, result.Name, nil
}

// ---------------------------------------------------------------------------
// Explore trips (cards)
// ---------------------------------------------------------------------------

// ListExploreTrips returns a paginated list of trending active packages.
func (s *HomeServices) ListExploreTrips(req bxapiReq.CustomerHomeExploreReq) (response.CustomerExploreTripsResult, error) {
	if global.GVA_DB == nil {
		return response.CustomerExploreTripsResult{}, errors.New("db not init")
	}
	result, _, err := cache.GetOrLoadJSON(context.Background(), ExploreTripsKey(req), sharedResponseTTL, func() (response.CustomerExploreTripsResult, error) {
		items, total, totalPage, err := s.listExploreTripsUncached(req)
		if err != nil {
			return response.CustomerExploreTripsResult{}, err
		}
		return response.CustomerExploreTripsResult{Items: items, Total: total, TotalPage: totalPage}, nil
	})
	return result, err
}

func (s *HomeServices) listExploreTripsUncached(req bxapiReq.CustomerHomeExploreReq) ([]response.CustomerExploreTripCard, int64, int64, error) {
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
	cards, err := s.buildExploreTripCards(packages)
	if err != nil {
		return nil, 0, 0, err
	}
	return cards, pr.Total, pr.TotalPage, nil
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
	result, _, err := cache.GetOrLoadJSON(context.Background(), DestinationsKey(page, pageSize), sharedResponseTTL, func() (cachedPage[response.CustomerDestinationCard], error) {
		paginateResult, err := s.listDestinationsUncached(page, pageSize)
		if err != nil {
			return cachedPage[response.CustomerDestinationCard]{}, err
		}
		items, ok := paginateResult.Data.([]response.CustomerDestinationCard)
		if !ok {
			return cachedPage[response.CustomerDestinationCard]{}, errors.New("unexpected destinations result type")
		}
		return cachedPage[response.CustomerDestinationCard]{
			Items: items, Total: paginateResult.Total, Page: paginateResult.Page,
			PageSize: paginateResult.PageSize, TotalPage: paginateResult.TotalPage,
		}, nil
	})
	if err != nil {
		return nil, err
	}
	return result.paginateResult(), nil
}

func (s *HomeServices) listDestinationsUncached(page, pageSize int) (*utils.PaginateResult, error) {
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
func (s *HomeServices) ListDeals(req bxapiReq.CustomerHomeDealsReq) (response.CustomerDealsResult, error) {
	if global.GVA_DB == nil {
		return response.CustomerDealsResult{}, errors.New("db not init")
	}
	result, _, err := cache.GetOrLoadJSON(context.Background(), DealsKey(req), sharedResponseTTL, func() (response.CustomerDealsResult, error) {
		items, total, totalPage, err := s.listDealsUncached(req)
		if err != nil {
			return response.CustomerDealsResult{}, err
		}
		return response.CustomerDealsResult{Items: items, Total: total, TotalPage: totalPage}, nil
	})
	return result, err
}

func (s *HomeServices) listDealsUncached(req bxapiReq.CustomerHomeDealsReq) ([]response.CustomerDealResult, int64, int64, error) {
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
	results, err := s.buildDealResults(packages)
	if err != nil {
		return nil, 0, 0, err
	}
	return results, pr.Total, pr.TotalPage, nil
}

// ---------------------------------------------------------------------------
// Top rated agencies
// ---------------------------------------------------------------------------

// ListTopRatedAgents returns active agents ranked by average review rating.
func (s *HomeServices) ListTopRatedAgents(page, pageSize int) (*utils.PaginateResult, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	result, _, err := cache.GetOrLoadJSON(context.Background(), TopRatedAgentsKey(page, pageSize), sharedResponseTTL, func() (cachedPage[response.CustomerAgentCard], error) {
		paginateResult, err := s.listTopRatedAgentsUncached(page, pageSize)
		if err != nil {
			return cachedPage[response.CustomerAgentCard]{}, err
		}
		items, ok := paginateResult.Data.([]response.CustomerAgentCard)
		if !ok {
			return cachedPage[response.CustomerAgentCard]{}, errors.New("unexpected top-rated agents result type")
		}
		return cachedPage[response.CustomerAgentCard]{
			Items: items, Total: paginateResult.Total, Page: paginateResult.Page,
			PageSize: paginateResult.PageSize, TotalPage: paginateResult.TotalPage,
		}, nil
	})
	if err != nil {
		return nil, err
	}
	return result.paginateResult(), nil
}

func (s *HomeServices) listTopRatedAgentsUncached(page, pageSize int) (*utils.PaginateResult, error) {
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
	if limit <= 0 {
		limit = 10
	}
	pool, _, err := cache.GetOrLoadJSON(context.Background(), RecommendedForYouKey(limit), sharedResponseTTL, func() ([]response.CustomerRecommendedResult, error) {
		var packages []catalog.Package
		if err := global.GVA_DB.Where("status = ?", constants.PackageStatusActive).
			Order("approved_at desc").Limit(limit).Find(&packages).Error; err != nil {
			return nil, err
		}
		agentIDs := make([]uint, 0, len(packages))
		agentSet := make(map[uint]struct{}, len(packages))
		packageIDs := make([]uint, 0, len(packages))
		for _, pkg := range packages {
			packageIDs = append(packageIDs, pkg.ID)
			if _, found := agentSet[pkg.AgentID]; !found {
				agentSet[pkg.AgentID] = struct{}{}
				agentIDs = append(agentIDs, pkg.AgentID)
			}
		}
		agentInfo, err := s.batchAgentNamesAndLogos(agentIDs)
		if err != nil {
			return nil, err
		}
		ratings, counts, err := s.batchPackageRatings(packageIDs)
		if err != nil {
			return nil, err
		}
		items := make([]response.CustomerRecommendedResult, 0, len(packages))
		for _, pkg := range packages {
			info := agentInfo[pkg.AgentID]
			items = append(items, response.CustomerRecommendedResult{
				Package: pkg, AgentName: info.Name, AgentLogo: info.Logo,
				Rating: ratings[pkg.ID], ReviewCount: counts[pkg.ID],
			})
		}
		return items, nil
	})
	if err != nil {
		return nil, err
	}

	packageIDs := make([]uint, 0, len(pool))
	for _, item := range pool {
		packageIDs = append(packageIDs, item.Package.ID)
	}
	var wishlistedIDs []uint
	if len(packageIDs) > 0 {
		var wishlist []engage.Wishlist
		if err := global.GVA_DB.Select("package_id").
			Where("user_id = ? AND package_id IN ?", userID, packageIDs).
			Find(&wishlist).Error; err != nil {
			return nil, err
		}
		for _, item := range wishlist {
			wishlistedIDs = append(wishlistedIDs, item.PackageID)
		}
	}
	applyWishlistFlags(pool, wishlistedIDs)
	return pool, nil
}

// ---------------------------------------------------------------------------
// Ads (home top banner)
// ---------------------------------------------------------------------------

// ActiveHomeAds returns active ad creatives for the home top banner placement with pagination.
func (s *HomeServices) ActiveHomeAds(page, pageSize int) (response.ActiveHomeAdsResult, error) {
	if global.GVA_DB == nil {
		return response.ActiveHomeAdsResult{}, errors.New("db not init")
	}
	result, _, err := cache.GetOrLoadJSON(context.Background(), HomeAdsKey(page, pageSize), sharedResponseTTL, func() (response.ActiveHomeAdsResult, error) {
		items, total, totalPage, err := s.activeHomeAdsUncached(page, pageSize)
		if err != nil {
			return response.ActiveHomeAdsResult{}, err
		}
		return response.ActiveHomeAdsResult{Items: items, Total: total, TotalPage: totalPage}, nil
	})
	return result, err
}

func (s *HomeServices) activeHomeAdsUncached(page, pageSize int) ([]response.AdBannerResult, int64, int64, error) {
	if global.GVA_DB == nil {
		return nil, 0, 0, errors.New("db not init")
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
		return nil, 0, 0, err
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
	return out, total, paginateResult.TotalPage, nil
}

// ---------------------------------------------------------------------------
// Private helpers — batch loading
// ---------------------------------------------------------------------------

// agentInfo holds the name and logo loaded for a single agent.
type agentInfo struct {
	Name string `json:"name"`
	Logo string `json:"logo"`
}

// batchAgentRatings loads avg rating and review count for multiple agents in one query.
func (s *HomeServices) batchAgentRatings(agentIDs []uint) (ratings map[uint]float64, counts map[uint]int, err error) {
	ratings = make(map[uint]float64)
	counts = make(map[uint]int)
	if len(agentIDs) == 0 {
		return ratings, counts, nil
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
		return nil, nil, err
	}
	for _, r := range rows {
		ratings[r.AgentID] = r.Avg
		counts[r.AgentID] = int(r.Cnt)
	}
	return ratings, counts, nil
}

// batchPackageRatings loads avg rating and review count for multiple packages in one query.
func (s *HomeServices) batchPackageRatings(pkgIDs []uint) (ratings map[uint]float64, counts map[uint]int, err error) {
	ratings = make(map[uint]float64)
	counts = make(map[uint]int)
	if len(pkgIDs) == 0 {
		return ratings, counts, nil
	}

	// Load review aggregates for the packages in one query.
	{
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
			return nil, nil, err
		}
		for _, r := range rows {
			ratings[r.PackageID] = r.Avg
			counts[r.PackageID] = int(r.Cnt)
		}
	}

	return ratings, counts, nil
}

// batchAgentNamesAndLogos loads company name and header image for multiple agents in one query.
func (s *HomeServices) batchAgentNamesAndLogos(agentIDs []uint) (map[uint]agentInfo, error) {
	out := make(map[uint]agentInfo, len(agentIDs))
	if len(agentIDs) == 0 {
		return out, nil
	}

	// Query all requested agents in one DB read.
	{
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
			return nil, err
		}
		for _, r := range rows {
			out[r.ID] = agentInfo{Name: r.CompanyName, Logo: r.HeaderImg}
		}
	}

	return out, nil
}

// ---------------------------------------------------------------------------
// Private helpers — result builders
// ---------------------------------------------------------------------------

func (s *HomeServices) buildAgentHits(agents []dxapi.DxAgentProfile) ([]response.CustomerAgentResult, error) {
	var agentIDs []uint
	for _, a := range agents {
		agentIDs = append(agentIDs, a.ID)
	}
	ratings, counts, err := s.batchAgentRatings(agentIDs)
	if err != nil {
		return nil, err
	}

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
	return out, nil
}

func (s *HomeServices) buildExploreHits(packages []catalog.Package) ([]response.CustomerExploreResult, error) {
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
	agentInfo, err := s.batchAgentNamesAndLogos(agentIDs)
	if err != nil {
		return nil, err
	}
	pkgRatings, pkgCounts, err := s.batchPackageRatings(pkgIDs)
	if err != nil {
		return nil, err
	}

	out := make([]response.CustomerExploreResult, 0, len(packages))
	for _, p := range packages {
		info, _ := agentInfo[p.AgentID]
		out = append(out, response.CustomerExploreResult{
			ID:              p.ID,
			PackageCode:     p.PackageCode,
			AgentID:         p.AgentID,
			CompanyName:     info.Name,
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
	return out, nil
}

func (s *HomeServices) buildExploreTripCards(packages []catalog.Package) ([]response.CustomerExploreTripCard, error) {
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
	agentInfo, err := s.batchAgentNamesAndLogos(agentIDs)
	if err != nil {
		return nil, err
	}
	pkgRatings, pkgCounts, err := s.batchPackageRatings(pkgIDs)
	if err != nil {
		return nil, err
	}

	out := make([]response.CustomerExploreTripCard, 0, len(packages))
	for _, p := range packages {
		info, _ := agentInfo[p.AgentID]
		out = append(out, response.CustomerExploreTripCard{
			ID:              p.ID,
			PackageCode:     p.PackageCode,
			AgentID:         p.AgentID,
			CompanyName:     info.Name,
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
	return out, nil
}

func (s *HomeServices) buildDealResults(packages []catalog.Package) ([]response.CustomerDealResult, error) {
	var agentIDs []uint
	agentSet := make(map[uint]bool)
	for _, p := range packages {
		if !agentSet[p.AgentID] {
			agentSet[p.AgentID] = true
			agentIDs = append(agentIDs, p.AgentID)
		}
	}
	agentInfo, err := s.batchAgentNamesAndLogos(agentIDs)
	if err != nil {
		return nil, err
	}

	out := make([]response.CustomerDealResult, 0, len(packages))
	for _, p := range packages {
		info, _ := agentInfo[p.AgentID]
		out = append(out, response.CustomerDealResult{
			Package:         p,
			AgentName:       info.Name,
			AgentLogo:       info.Logo,
			DiscountPercent: p.DiscountPercent,
			ViewCount:       0, // view count not tracked yet; field reserved
		})
	}
	return out, nil
}

func (s *HomeServices) truncateDesc(text string, max int) string {
	runes := []rune(text)
	if len(runes) <= max {
		return text
	}
	return string(runes[:max]) + "…"
}
