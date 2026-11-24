package bxapi

import (
	"context"
	"errors"

	"github.com/mo7amed4522/tayara_api/constants"
	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/model/bxapi/request"
	bxapiResp "github.com/mo7amed4522/tayara_api/model/bxapi/response"
	catalogModel "github.com/mo7amed4522/tayara_api/model/catalog"
	engageModel "github.com/mo7amed4522/tayara_api/model/engage"
	engageSvc "github.com/mo7amed4522/tayara_api/service/engage"
	"github.com/mo7amed4522/tayara_api/utils"
	"github.com/mo7amed4522/tayara_api/utils/cache"
)

// ExploreServices provides explore topic operations for public/agent APIs
type ExploreServices struct{}

// ============================================================================
// Public (User-facing) Methods
// ============================================================================

// ListExploreTopics returns paginated active explore topics
func (s *ExploreServices) ListExploreTopics(req request.CustomerExploreTopicsReq) (*utils.PaginateResult, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	// Set defaults for pagination
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	result, _, err := cache.GetOrLoadJSON(context.Background(), ExploreTopicsKey(req.Page, req.PageSize), sharedResponseTTL, func() (cachedPage[bxapiResp.CustomerExploreTopicCard], error) {
		paginateResult, err := s.listExploreTopicsUncached(req)
		if err != nil {
			return cachedPage[bxapiResp.CustomerExploreTopicCard]{}, err
		}
		items, ok := paginateResult.Data.([]bxapiResp.CustomerExploreTopicCard)
		if !ok {
			return cachedPage[bxapiResp.CustomerExploreTopicCard]{}, errors.New("unexpected explore topics result type")
		}
		return cachedPage[bxapiResp.CustomerExploreTopicCard]{
			Items: items, Total: paginateResult.Total, Page: paginateResult.Page,
			PageSize: paginateResult.PageSize, TotalPage: paginateResult.TotalPage,
		}, nil
	})
	if err != nil {
		return nil, err
	}
	return result.paginateResult(), nil
}

func (s *ExploreServices) listExploreTopicsUncached(req request.CustomerExploreTopicsReq) (*utils.PaginateResult, error) {
	topicService := &engageSvc.ExploreTopicServices{}

	// Build count query
	countDB := global.GVA_DB.Model(&engageModel.ExploreTopic{}).
		Where("is_active = ? AND deleted_at IS NULL", true)

	// Build data query
	queryDB := global.GVA_DB.Model(&engageModel.ExploreTopic{}).
		Where("is_active = ? AND deleted_at IS NULL", true).
		Order("sort_order asc, id asc")

	var topics []engageModel.ExploreTopic
	paginateResult, err := utils.PaginateWithCustomQuery(countDB, queryDB, req.Page, req.PageSize, &topics)
	if err != nil {
		return nil, err
	}

	// Get package counts for each topic
	topicIDs := make([]uint, len(topics))
	for i, t := range topics {
		topicIDs[i] = t.ID
	}
	counts, err := topicService.GetPackageCountsByTopicE(topicIDs)
	if err != nil {
		return nil, err
	}

	// Convert to bxapi response format
	cards := make([]bxapiResp.CustomerExploreTopicCard, len(topics))
	for i, t := range topics {
		cards[i] = bxapiResp.CustomerExploreTopicCard{
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

// GetExploreTopicDetail returns a topic with featured packages
func (s *ExploreServices) GetExploreTopicDetail(code string) (*bxapiResp.CustomerExploreTopicDetail, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	detail, _, err := cache.GetOrLoadJSON(context.Background(), ExploreTopicDetailKey(code), sharedResponseTTL, func() (*bxapiResp.CustomerExploreTopicDetail, error) {
		return s.getExploreTopicDetailUncached(code)
	})
	return detail, err
}

func (s *ExploreServices) getExploreTopicDetailUncached(code string) (*bxapiResp.CustomerExploreTopicDetail, error) {
	topicService := &engageSvc.ExploreTopicServices{}
	detail, err := topicService.GetTopicDetail(code)
	if err != nil {
		return nil, err
	}

	// Convert featured packages to bxapi response format
	featured := make([]bxapiResp.CustomerExploreTripCard, len(detail.FeaturedPackages))
	copy(featured, detail.FeaturedPackages)

	return &bxapiResp.CustomerExploreTopicDetail{
		CustomerExploreTopicCard: bxapiResp.CustomerExploreTopicCard{
			ID:           detail.ID,
			Code:         detail.Code,
			Name:         detail.Name,
			NameAr:       detail.NameAr,
			Description:  detail.Description,
			ImageURL:     detail.ImageURL,
			IconURL:      detail.IconURL,
			PackageCount: detail.PackageCount,
		},
		FeaturedPackages: featured,
	}, nil
}

// GetExploreTopicPackages returns paginated packages for a topic
func (s *ExploreServices) GetExploreTopicPackages(code string, req request.CustomerExploreTopicPackagesReq) (*bxapiResp.CustomerExploreTopicPackagesResult, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	result, _, err := cache.GetOrLoadJSON(context.Background(), ExploreTopicPackagesKey(code, req), sharedResponseTTL, func() (*bxapiResp.CustomerExploreTopicPackagesResult, error) {
		return s.getExploreTopicPackagesUncached(code, req)
	})
	return result, err
}

func (s *ExploreServices) getExploreTopicPackagesUncached(code string, req request.CustomerExploreTopicPackagesReq) (*bxapiResp.CustomerExploreTopicPackagesResult, error) {
	topicService := &engageSvc.ExploreTopicServices{}
	// Convert to internal request format
	internalReq := request.CustomerHomeExploreReq{
		PageInfo: req.PageInfo,
	}

	result, err := topicService.GetTopicPackages(code, internalReq)
	if err != nil {
		return nil, err
	}

	// Convert packages to bxapi response format
	packages := make([]bxapiResp.CustomerExploreTripCard, len(result.Packages))
	copy(packages, result.Packages)

	return &bxapiResp.CustomerExploreTopicPackagesResult{
		Topic: bxapiResp.CustomerExploreTopicCard{
			ID:           result.Topic.ID,
			Code:         result.Topic.Code,
			Name:         result.Topic.Name,
			NameAr:       result.Topic.NameAr,
			Description:  result.Topic.Description,
			ImageURL:     result.Topic.ImageURL,
			IconURL:      result.Topic.IconURL,
			PackageCount: result.Topic.PackageCount,
		},
		Packages: packages,
		Total:    result.Total,
		Page:     result.Page,
		PageSize: result.PageSize,
	}, nil
}

// ============================================================================
// Agent-facing Methods
// ============================================================================

// AgentGetMyTopics returns paginated topics with agent's package counts
func (s *ExploreServices) AgentGetMyTopics(agentID uint, page, pageSize int) (*utils.PaginateResult, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}

	// Set defaults for pagination
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	result, _, err := cache.GetOrLoadJSON(context.Background(), AgentGetMyTopicsKey(agentID, page, pageSize), agentScopedTTL, func() (cachedPage[bxapiResp.AgentExploreTopicCard], error) {
		paginateResult, err := s.agentGetMyTopicsUncached(agentID, page, pageSize)
		if err != nil {
			return cachedPage[bxapiResp.AgentExploreTopicCard]{}, err
		}
		items, ok := paginateResult.Data.([]bxapiResp.AgentExploreTopicCard)
		if !ok {
			return cachedPage[bxapiResp.AgentExploreTopicCard]{}, errors.New("unexpected agent topics result type")
		}
		return cachedPage[bxapiResp.AgentExploreTopicCard]{
			Items: items, Total: paginateResult.Total, Page: paginateResult.Page,
			PageSize: paginateResult.PageSize, TotalPage: paginateResult.TotalPage,
		}, nil
	})
	if err != nil {
		return nil, err
	}
	return result.paginateResult(), nil
}

func (s *ExploreServices) agentGetMyTopicsUncached(agentID uint, page, pageSize int) (*utils.PaginateResult, error) {
	topicService := &engageSvc.ExploreTopicServices{}

	// Get all active topics (we need to count them for pagination)
	// Build count query - count all active topics
	countDB := global.GVA_DB.Model(&engageModel.ExploreTopic{}).
		Where("is_active = ? AND deleted_at IS NULL", true)

	// Build data query - get all active topics ordered
	queryDB := global.GVA_DB.Model(&engageModel.ExploreTopic{}).
		Where("is_active = ? AND deleted_at IS NULL", true).
		Order("sort_order asc, id asc")

	var topics []engageModel.ExploreTopic
	paginateResult, err := utils.PaginateWithCustomQuery(countDB, queryDB, page, pageSize, &topics)
	if err != nil {
		return nil, err
	}

	// Get agent's package IDs
	var agentPackageIDs []uint
	err = global.GVA_DB.Model(&catalogModel.Package{}).
		Where("agent_id = ? AND status = ?", agentID, constants.PackageStatusActive).
		Pluck("id", &agentPackageIDs).Error
	if err != nil {
		return nil, err
	}

	// Count packages per topic for this agent
	var packageTopics []engageModel.PackageTopic
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
	totalCounts, err := topicService.GetPackageCountsByTopicE(topicIDs)
	if err != nil {
		return nil, err
	}

	// Build response cards with both total and agent-specific counts
	cards := make([]bxapiResp.AgentExploreTopicCard, len(topics))
	for i, t := range topics {
		cards[i] = bxapiResp.AgentExploreTopicCard{
			CustomerExploreTopicCard: bxapiResp.CustomerExploreTopicCard{
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

// AgentAddPackageToTopics adds a package to topics
func (s *ExploreServices) AgentAddPackageToTopics(agentID, packageID uint, topicIDs []uint) error {
	if global.GVA_DB == nil {
		return errors.New("db not init")
	}

	topicService := &engageSvc.ExploreTopicServices{}
	return topicService.AgentAddPackageToTopic(agentID, packageID, topicIDs)
}

// AgentRemovePackageFromTopic removes a package from a topic
func (s *ExploreServices) AgentRemovePackageFromTopic(agentID, packageID, topicID uint) error {
	if global.GVA_DB == nil {
		return errors.New("db not init")
	}

	topicService := &engageSvc.ExploreTopicServices{}
	return topicService.AgentRemovePackageFromTopic(agentID, packageID, topicID)
}

// AgentGetPackageTopics returns topic IDs for a package
func (s *ExploreServices) AgentGetPackageTopics(agentID, packageID uint) ([]uint, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	result, _, err := cache.GetOrLoadJSON(context.Background(), AgentGetPackageTopicsKey(agentID, packageID), agentScopedTTL, func() ([]uint, error) {
		topicService := &engageSvc.ExploreTopicServices{}
		return topicService.AgentGetPackageTopics(agentID, packageID)
	})
	return result, err
}
