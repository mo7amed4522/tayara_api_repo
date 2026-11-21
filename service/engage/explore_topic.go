package engage

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/mo7amed4522/tayara_api/constants"
	"github.com/mo7amed4522/tayara_api/global"
	bxapiReq "github.com/mo7amed4522/tayara_api/model/bxapi/request"
	bxapiResp "github.com/mo7amed4522/tayara_api/model/bxapi/response"
	"github.com/mo7amed4522/tayara_api/model/catalog"
	"github.com/mo7amed4522/tayara_api/model/common/request"
	"github.com/mo7amed4522/tayara_api/model/dxapi"
	"github.com/mo7amed4522/tayara_api/model/engage"
	engageReq "github.com/mo7amed4522/tayara_api/model/engage/request"
	engageResp "github.com/mo7amed4522/tayara_api/model/engage/response"
	"github.com/mo7amed4522/tayara_api/utils"
	"github.com/mo7amed4522/tayara_api/utils/cache"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ExploreTopicServices handles explore topic operations
type ExploreTopicServices struct{}

// ============================================================================
// Public (User-facing) Methods
// ============================================================================

// ListTopics returns a paginated list of active explore topics with package counts.
func (s *ExploreTopicServices) ListTopics(page, pageSize int) (*utils.PaginateResult, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}

	var topics []engage.ExploreTopic
	countDB := global.GVA_DB.Model(&engage.ExploreTopic{}).
		Where("is_active = ? AND deleted_at IS NULL", true)
	queryDB := global.GVA_DB.Model(&engage.ExploreTopic{}).
		Where("is_active = ? AND deleted_at IS NULL", true).
		Order("sort_order asc, id asc")
	paginateResult, err := utils.PaginateWithCustomQuery(countDB, queryDB, page, pageSize, &topics)
	if err != nil {
		return nil, err
	}

	// Get package counts for each topic
	topicIDs := make([]uint, len(topics))
	for i, t := range topics {
		topicIDs[i] = t.ID
	}

	counts := s.GetPackageCountsByTopic(topicIDs)

	// Build response cards
	cards := make([]engageResp.ExploreTopicCard, len(topics))
	for i, t := range topics {
		cards[i] = engageResp.ExploreTopicCard{
			ID:           t.ID,
			Code:         t.Code,
			Name:         t.Name,
			NameAr:       t.NameAr,
			Description:  t.Description,
			ImageURL:     t.ImageURL,
			IconURL:      t.IconURL,
			PackageCount: counts[t.ID],
		}
	}

	paginateResult.Data = cards
	return paginateResult, nil
}

// GetTopicDetail returns topic metadata with featured packages
func (s *ExploreTopicServices) GetTopicDetail(topicCode string) (*engageResp.ExploreTopicDetail, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}

	// Try to get cached topic by code
	cacheKey := "engage:explore:topic:code:" + topicCode
	if val, found, _ := cache.Get(cacheKey); found {
		if topic, ok := val.(*engage.ExploreTopic); ok {
			if !topic.IsActive {
				return nil, errors.New("topic not available")
			}

			// Get featured packages (first 6 active packages in this topic)
			featuredPackages, err := s.getTopicPackages(topic.ID, 1, 6, true)
			if err != nil {
				return nil, err
			}

			// Get package count
			counts := s.GetPackageCountsByTopic([]uint{topic.ID})

			return &engageResp.ExploreTopicDetail{
				ExploreTopicCard: engageResp.ExploreTopicCard{
					ID:           topic.ID,
					Code:         topic.Code,
					Name:         topic.Name,
					NameAr:       topic.NameAr,
					Description:  topic.Description,
					ImageURL:     topic.ImageURL,
					IconURL:      topic.IconURL,
					PackageCount: counts[topic.ID],
				},
				FeaturedPackages: featuredPackages,
			}, nil
		}
	}

	topic, err := engage.GetTopicByCode(topicCode)
	if err != nil {
		return nil, errors.New("topic not found")
	}
	if !topic.IsActive {
		return nil, errors.New("topic not available")
	}

	// Cache the topic
	cache.Set(cacheKey, topic, 128, 10*time.Minute)

	// Get featured packages (first 6 active packages in this topic)
	featuredPackages, err := s.getTopicPackages(topic.ID, 1, 6, true)
	if err != nil {
		return nil, err
	}

	// Get package count
	counts := s.GetPackageCountsByTopic([]uint{topic.ID})

	return &engageResp.ExploreTopicDetail{
		ExploreTopicCard: engageResp.ExploreTopicCard{
			ID:           topic.ID,
			Code:         topic.Code,
			Name:         topic.Name,
			NameAr:       topic.NameAr,
			Description:  topic.Description,
			ImageURL:     topic.ImageURL,
			IconURL:      topic.IconURL,
			PackageCount: counts[topic.ID],
		},
		FeaturedPackages: featuredPackages,
	}, nil
}

// GetTopicPackages returns paginated packages for a specific topic
func (s *ExploreTopicServices) GetTopicPackages(topicCode string, req bxapiReq.CustomerHomeExploreReq) (*engageResp.ExploreTopicPackagesResult, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}

	topic, err := engage.GetTopicByCode(topicCode)
	if err != nil {
		return nil, errors.New("topic not found")
	}
	if !topic.IsActive {
		return nil, errors.New("topic not available")
	}

	packages, total, err := s.getTopicPackagesPaginated(topic.ID, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}

	counts := s.GetPackageCountsByTopic([]uint{topic.ID})

	return &engageResp.ExploreTopicPackagesResult{
		Topic: engageResp.ExploreTopicCard{
			ID:           topic.ID,
			Code:         topic.Code,
			Name:         topic.Name,
			NameAr:       topic.NameAr,
			Description:  topic.Description,
			ImageURL:     topic.ImageURL,
			IconURL:      topic.IconURL,
			PackageCount: counts[topic.ID],
		},
		Packages: packages,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// ============================================================================
// Agent-facing Methods
// ============================================================================

// AgentGetMyTopics returns a paginated list of topics with the agent's package counts.
func (s *ExploreTopicServices) AgentGetMyTopics(agentID uint, page, pageSize int) (*utils.PaginateResult, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}

	var topics []engage.ExploreTopic
	countDB := global.GVA_DB.Model(&engage.ExploreTopic{}).
		Where("is_active = ? AND deleted_at IS NULL", true)
	queryDB := global.GVA_DB.Model(&engage.ExploreTopic{}).
		Where("is_active = ? AND deleted_at IS NULL", true).
		Order("sort_order asc, id asc")
	paginateResult, err := utils.PaginateWithCustomQuery(countDB, queryDB, page, pageSize, &topics)
	if err != nil {
		return nil, err
	}

	// Get agent's package IDs
	var agentPackageIDs []uint
	err = global.GVA_DB.Model(&catalog.Package{}).
		Where("agent_id = ? AND status = ?", agentID, constants.PackageStatusActive).
		Pluck("id", &agentPackageIDs).Error
	if err != nil {
		return nil, err
	}

	// Get topic associations for agent's packages
	var packageTopics []engage.PackageTopic
	if len(agentPackageIDs) > 0 {
		err = global.GVA_DB.Where("package_id IN ?", agentPackageIDs).Find(&packageTopics).Error
		if err != nil {
			return nil, err
		}
	}

	// Count packages per topic for this agent
	agentTopicCounts := make(map[uint]int)
	for _, pt := range packageTopics {
		agentTopicCounts[pt.TopicID]++
	}

	// Get total package counts per topic (for all agents)
	topicIDs := make([]uint, len(topics))
	for i, t := range topics {
		topicIDs[i] = t.ID
	}
	totalCounts := s.GetPackageCountsByTopic(topicIDs)

	// Build response
	cards := make([]engageResp.AgentTopicCard, len(topics))
	for i, t := range topics {
		cards[i] = engageResp.AgentTopicCard{
			ExploreTopicCard: engageResp.ExploreTopicCard{
				ID:           t.ID,
				Code:         t.Code,
				Name:         t.Name,
				NameAr:       t.NameAr,
				Description:  t.Description,
				ImageURL:     t.ImageURL,
				IconURL:      t.IconURL,
				PackageCount: totalCounts[t.ID],
			},
			MyPackageCount: agentTopicCounts[t.ID],
		}
	}

	paginateResult.Data = cards
	return paginateResult, nil
}

// AgentAddPackageToTopic associates a package with one or more topics
func (s *ExploreTopicServices) AgentAddPackageToTopic(agentID, packageID uint, topicIDs []uint) error {
	if global.GVA_DB == nil {
		return errors.New("db not init")
	}

	// Verify package belongs to agent and is active
	var pkg catalog.Package
	err := global.GVA_DB.Where("id = ? AND agent_id = ? AND status = ?", packageID, agentID, constants.PackageStatusActive).First(&pkg).Error
	if err != nil {
		return errors.New("package not found or not authorized")
	}

	// Verify topics exist and are active
	var validTopics []engage.ExploreTopic
	err = global.GVA_DB.Where("id IN ? AND is_active = ? AND deleted_at IS NULL", topicIDs, true).Find(&validTopics).Error
	if err != nil {
		return err
	}
	if len(validTopics) != len(topicIDs) {
		return errors.New("one or more topics not found or inactive")
	}

	// Create associations (ignore duplicates)
	now := global.GVA_DB.NowFunc()
	associations := make([]engage.PackageTopic, 0, len(topicIDs))
	for _, topicID := range topicIDs {
		associations = append(associations, engage.PackageTopic{
			PackageID: packageID,
			TopicID:   topicID,
			AddedBy:   agentID,
			AddedAt:   now,
		})
	}

	// Use ON CONFLICT DO NOTHING (PostgreSQL) or INSERT IGNORE (MySQL)
	return global.GVA_DB.Clauses(clause.OnConflict{DoNothing: true}).Create(&associations).Error
}

// AgentRemovePackageFromTopic removes package-topic association
func (s *ExploreTopicServices) AgentRemovePackageFromTopic(agentID, packageID, topicID uint) error {
	if global.GVA_DB == nil {
		return errors.New("db not init")
	}

	// Verify package belongs to agent
	var pkg catalog.Package
	err := global.GVA_DB.Where("id = ? AND agent_id = ?", packageID, agentID).First(&pkg).Error
	if err != nil {
		return errors.New("package not found or not authorized")
	}

	return global.GVA_DB.Where("package_id = ? AND topic_id = ?", packageID, topicID).
		Delete(&engage.PackageTopic{}).Error
}

// AgentGetPackageTopics returns topic IDs associated with a package
func (s *ExploreTopicServices) AgentGetPackageTopics(agentID, packageID uint) ([]uint, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}

	// Verify package belongs to agent
	var pkg catalog.Package
	err := global.GVA_DB.Where("id = ? AND agent_id = ?", packageID, agentID).First(&pkg).Error
	if err != nil {
		return nil, errors.New("package not found or not authorized")
	}

	var topicIDs []uint
	err = global.GVA_DB.Model(&engage.PackageTopic{}).
		Where("package_id = ?", packageID).
		Pluck("topic_id", &topicIDs).Error
	return topicIDs, err
}

// ============================================================================
// Admin-facing Methods
// ============================================================================

// AdminListTopics returns paginated list of all topics (including inactive)
func (s *ExploreTopicServices) AdminListTopics(req engageReq.AdminExploreTopicPageReq) (*utils.PaginateResult, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}

	db := global.GVA_DB.Model(&engage.ExploreTopic{}).Where("deleted_at IS NULL")
	if req.IsActive != nil {
		db = db.Where("is_active = ?", *req.IsActive)
	}
	if req.Keyword != "" {
		kw := "%" + strings.TrimSpace(req.Keyword) + "%"
		db = db.Where("name LIKE ? OR name_ar LIKE ? OR code LIKE ?", kw, kw, kw)
	}

	var topics []engage.ExploreTopic
	countDB := db.Session(&gorm.Session{})
	queryDB := db.Order("sort_order asc, id asc")

	return utils.PaginateWithCustomQuery(countDB, queryDB, req.Page, req.PageSize, &topics)
}

// AdminGetTopic returns a single topic by ID (admin can see inactive)
func (s *ExploreTopicServices) AdminGetTopic(id uint) (*engage.ExploreTopic, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}

	var topic engage.ExploreTopic
	err := global.GVA_DB.Where("id = ? AND deleted_at IS NULL", id).First(&topic).Error
	if err != nil {
		return nil, errors.New("topic not found")
	}
	return &topic, nil
}

// AdminCreateTopic creates a new explore topic
func (s *ExploreTopicServices) AdminCreateTopic(req engageReq.ExploreTopicReq) (*engage.ExploreTopic, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}

	// Check code uniqueness
	var existing engage.ExploreTopic
	if err := global.GVA_DB.Where("code = ?", req.Code).First(&existing).Error; err == nil {
		return nil, errors.New("topic code already exists")
	}

	topic := &engage.ExploreTopic{
		Code:        req.Code,
		Name:        req.Name,
		NameAr:      req.NameAr,
		Description: req.Description,
		ImageURL:    req.ImageURL,
		IconURL:     req.IconURL,
		SortOrder:   req.SortOrder,
		IsActive:    req.IsActive,
	}

	err := global.GVA_DB.Create(topic).Error
	return topic, err
}

// AdminUpdateTopic updates an existing topic
func (s *ExploreTopicServices) AdminUpdateTopic(id uint, req engageReq.ExploreTopicReq) (*engage.ExploreTopic, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}

	var topic engage.ExploreTopic
	if err := global.GVA_DB.Where("id = ? AND deleted_at IS NULL", id).First(&topic).Error; err != nil {
		return nil, errors.New("topic not found")
	}

	// Check code uniqueness (excluding self)
	var existing engage.ExploreTopic
	if err := global.GVA_DB.Where("code = ? AND id <> ?", req.Code, id).First(&existing).Error; err == nil {
		return nil, errors.New("topic code already exists")
	}

	updates := map[string]interface{}{
		"code":        req.Code,
		"name":        req.Name,
		"name_ar":     req.NameAr,
		"description": req.Description,
		"image_url":   req.ImageURL,
		"icon_url":    req.IconURL,
		"sort_order":  req.SortOrder,
		"is_active":   req.IsActive,
		"updated_at":  global.GVA_DB.NowFunc(),
	}

	err := global.GVA_DB.Model(&topic).Updates(updates).Error
	if err != nil {
		return nil, err
	}

	// Reload
	global.GVA_DB.Where("id = ?", id).First(&topic)
	return &topic, nil
}

// AdminDeleteTopic soft deletes a topic
func (s *ExploreTopicServices) AdminDeleteTopic(id uint) error {
	if global.GVA_DB == nil {
		return errors.New("db not init")
	}

	// Check if topic has packages
	var count int64
	global.GVA_DB.Model(&engage.PackageTopic{}).Where("topic_id = ?", id).Count(&count)
	if count > 0 {
		return errors.New("cannot delete topic with associated packages; remove packages first")
	}

	return global.GVA_DB.Where("id = ?", id).Delete(&engage.ExploreTopic{}).Error
}

// AdminReorderTopics bulk updates sort order for multiple topics
func (s *ExploreTopicServices) AdminReorderTopics(orders []engageReq.TopicOrderReq) error {
	if global.GVA_DB == nil {
		return errors.New("db not init")
	}

	return global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		for _, o := range orders {
			if err := tx.Model(&engage.ExploreTopic{}).
				Where("id = ?", o.ID).
				Update("sort_order", o.SortOrder).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// AdminManageTopicPackages bulk adds/removes packages for a topic
func (s *ExploreTopicServices) AdminManageTopicPackages(topicID uint, packageIDs []uint, replace bool) error {
	if global.GVA_DB == nil {
		return errors.New("db not init")
	}

	// Verify topic exists
	var topic engage.ExploreTopic
	if err := global.GVA_DB.Where("id = ? AND deleted_at IS NULL", topicID).First(&topic).Error; err != nil {
		return errors.New("topic not found")
	}

	// Verify all packages exist and are active
	var validCount int64
	global.GVA_DB.Model(&catalog.Package{}).
		Where("id IN ? AND status = ?", packageIDs, constants.PackageStatusActive).
		Count(&validCount)
	if int(validCount) != len(packageIDs) {
		return errors.New("one or more packages not found or not active")
	}

	return global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		if replace {
			// Remove all existing associations for this topic
			if err := tx.Where("topic_id = ?", topicID).Delete(&engage.PackageTopic{}).Error; err != nil {
				return err
			}
		}

		// Add new associations
		now := tx.NowFunc()
		associations := make([]engage.PackageTopic, 0, len(packageIDs))
		for _, pkgID := range packageIDs {
			// Check if already exists (if not replace mode)
			if !replace {
				var existing engage.PackageTopic
				if tx.Where("package_id = ? AND topic_id = ?", pkgID, topicID).First(&existing).Error == nil {
					continue // Skip existing
				}
			}
			associations = append(associations, engage.PackageTopic{
				PackageID: pkgID,
				TopicID:   topicID,
				AddedBy:   0, // Admin ID would be passed in real implementation
				AddedAt:   now,
			})
		}

		if len(associations) > 0 {
			return tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&associations).Error
		}
		return nil
	})
}

// AdminGetTopicPackages returns packages associated with a topic (for admin management)
func (s *ExploreTopicServices) AdminGetTopicPackages(topicID uint, page, pageSize int) (*utils.PaginateResult, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}

	// Verify topic exists
	var topic engage.ExploreTopic
	if err := global.GVA_DB.Where("id = ? AND deleted_at IS NULL", topicID).First(&topic).Error; err != nil {
		return nil, errors.New("topic not found")
	}

	// Build query joining packages with topic association
	type PackageWithTopic struct {
		catalog.Package
		AddedAt time.Time `gorm:"column:added_at"`
	}

	countSQL := global.GVA_DB.Model(&engage.PackageTopic{}).
		Where("topic_id = ?", topicID).
		Joins("INNER JOIN catalog_package ON catalog_package.id = engage_package_topic.package_id").
		Where("catalog_package.status = ?", constants.PackageStatusActive)

	querySQL := global.GVA_DB.Model(&engage.PackageTopic{}).
		Select("catalog_package.*, engage_package_topic.added_at").
		Where("engage_package_topic.topic_id = ?", topicID).
		Joins("INNER JOIN catalog_package ON catalog_package.id = engage_package_topic.package_id").
		Where("catalog_package.status = ?", constants.PackageStatusActive).
		Order("engage_package_topic.added_at desc")

	var packages []PackageWithTopic
	return utils.PaginateWithCustomQuery(countSQL, querySQL, page, pageSize, &packages)
}

// ============================================================================
// Private Helper Methods
// ============================================================================

// GetPackageCountsByTopic returns package counts for multiple topics (public for admin use)
func (s *ExploreTopicServices) GetPackageCountsByTopic(topicIDs []uint) map[uint]int {
	counts := make(map[uint]int)
	if len(topicIDs) == 0 {
		return counts
	}

	// Check cache for each topic
	var missingIDs []uint
	for _, topicID := range topicIDs {
		cacheKey := fmt.Sprintf("engage:explore:topic:%d:pkgcount", topicID)
		if val, found, _ := cache.Get(cacheKey); found {
			if count, ok := val.(int); ok {
				counts[topicID] = count
				continue
			}
		}
		missingIDs = append(missingIDs, topicID)
	}

	// Query DB only for missing IDs
	if len(missingIDs) > 0 {
		type countRow struct {
			TopicID uint `gorm:"column:topic_id"`
			Count   int  `gorm:"column:cnt"`
		}
		var rows []countRow
		global.GVA_DB.Model(&engage.PackageTopic{}).
			Select("topic_id, COUNT(*) as cnt").
			Where("topic_id IN ?", missingIDs).
			Joins("INNER JOIN catalog_package ON catalog_package.id = engage_package_topic.package_id").
			Where("catalog_package.status = ?", constants.PackageStatusActive).
			Group("topic_id").
			Scan(&rows)

		for _, r := range rows {
			counts[r.TopicID] = r.Count
			// Cache the result
			cacheKey := fmt.Sprintf("engage:explore:topic:%d:pkgcount", r.TopicID)
			cache.Set(cacheKey, r.Count, 8, 10*time.Minute)
		}
	}

	return counts
}

// getTopicPackages returns packages for a topic (limited, for featured)
func (s *ExploreTopicServices) getTopicPackages(topicID uint, page, pageSize int, featuredOnly bool) ([]bxapiResp.CustomerExploreTripCard, error) {
	var packages []catalog.Package
	query := global.GVA_DB.Model(&catalog.Package{}).
		Where("status = ?", constants.PackageStatusActive).
		Joins("INNER JOIN engage_package_topic ON engage_package_topic.package_id = catalog_package.id").
		Where("engage_package_topic.topic_id = ?", topicID).
		Order("approved_at desc")

	if featuredOnly {
		query = query.Limit(pageSize)
	} else {
		pageInfo := request.PageInfo{Page: page, PageSize: pageSize}
		query = query.Scopes(pageInfo.Paginate())
	}

	err := query.Find(&packages).Error
	if err != nil {
		return nil, err
	}

	// Build response cards with agent info and ratings
	return s.buildExploreTripCards(packages), nil
}

// getTopicPackagesPaginated returns paginated packages for a topic
func (s *ExploreTopicServices) getTopicPackagesPaginated(topicID uint, page, pageSize int) ([]bxapiResp.CustomerExploreTripCard, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	// Count query
	countDB := global.GVA_DB.Model(&engage.PackageTopic{}).
		Where("topic_id = ?", topicID).
		Joins("INNER JOIN catalog_package ON catalog_package.id = engage_package_topic.package_id").
		Where("catalog_package.status = ?", constants.PackageStatusActive)

	// Data query
	queryDB := global.GVA_DB.Model(&engage.PackageTopic{}).
		Select("catalog_package.*").
		Where("engage_package_topic.topic_id = ?", topicID).
		Joins("INNER JOIN catalog_package ON catalog_package.id = engage_package_topic.package_id").
		Where("catalog_package.status = ?", constants.PackageStatusActive).
		Order("catalog_package.approved_at desc")

	var packages []catalog.Package
	paginateResult, err := utils.PaginateWithCustomQuery(countDB, queryDB, page, pageSize, &packages)
	if err != nil {
		return nil, 0, err
	}

	return s.buildExploreTripCards(packages), paginateResult.Total, nil
}

// buildExploreTripCards converts packages to explore trip cards with agent info and ratings
func (s *ExploreTopicServices) buildExploreTripCards(packages []catalog.Package) []bxapiResp.CustomerExploreTripCard {
	if len(packages) == 0 {
		return []bxapiResp.CustomerExploreTripCard{}
	}

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

	// Batch load agent info
	agentInfo := s.batchAgentNamesAndLogos(agentIDs)
	// Batch load package ratings
	pkgRatings, pkgCounts := s.batchPackageRatings(pkgIDs)

	out := make([]bxapiResp.CustomerExploreTripCard, 0, len(packages))
	for _, p := range packages {
		info, _ := agentInfo[p.AgentID]
		out = append(out, bxapiResp.CustomerExploreTripCard{
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

// batchAgentNamesAndLogos loads company name and header image for multiple agents
func (s *ExploreTopicServices) batchAgentNamesAndLogos(agentIDs []uint) map[uint]struct{ name, logo string } {
	out := make(map[uint]struct{ name, logo string }, len(agentIDs))
	if len(agentIDs) == 0 {
		return out
	}
	type row struct {
		ID          uint   `gorm:"column:id"`
		CompanyName string `gorm:"column:company_name"`
		HeaderImg   string `gorm:"column:header_img"`
	}
	var rows []row
	global.GVA_DB.Model(&dxapi.DxAgentProfile{}).
		Where("id IN ?", agentIDs).
		Select("id, company_name, header_img").
		Scan(&rows)
	for _, r := range rows {
		out[r.ID] = struct{ name, logo string }{name: r.CompanyName, logo: r.HeaderImg}
	}
	return out
}

// batchPackageRatings loads avg rating and review count for multiple packages
func (s *ExploreTopicServices) batchPackageRatings(pkgIDs []uint) (map[uint]float64, map[uint]int) {
	ratings := make(map[uint]float64)
	counts := make(map[uint]int)
	if len(pkgIDs) == 0 {
		return ratings, counts
	}
	type row struct {
		PackageID uint    `gorm:"column:package_id"`
		Avg       float64 `gorm:"column:avg"`
		Cnt       int64   `gorm:"column:cnt"`
	}
	var rows []row
	global.GVA_DB.Model(&engage.Review{}).
		Where("package_id IN ?", pkgIDs).
		Where("review_status = ?", constants.ReviewStatusActive).
		Select("package_id, ROUND(AVG(rating), 2) AS avg, COUNT(*) AS cnt").
		Group("package_id").
		Scan(&rows)
	for _, r := range rows {
		ratings[r.PackageID] = r.Avg
		counts[r.PackageID] = int(r.Cnt)
	}
	return ratings, counts
}
