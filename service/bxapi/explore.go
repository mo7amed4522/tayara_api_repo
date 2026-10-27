package bxapi

import (
	"errors"

	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/model/bxapi/request"
	bxapiResp "github.com/mo7amed4522/tayara_api/model/bxapi/response"
	"github.com/mo7amed4522/tayara_api/service/engage"
)

// ExploreServices provides explore topic operations for public/agent APIs
type ExploreServices struct{}

// ============================================================================
// Public (User-facing) Methods
// ============================================================================

// ListExploreTopics returns all active explore topics
func (s *ExploreServices) ListExploreTopics(req request.CustomerExploreTopicsReq) ([]bxapiResp.CustomerExploreTopicCard, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}

	topicService := &engage.ExploreTopicServices{}
	topics, err := topicService.ListTopics()
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
			PackageCount: t.PackageCount,
		}
	}

	return cards, nil
}

// GetExploreTopicDetail returns a topic with featured packages
func (s *ExploreServices) GetExploreTopicDetail(code string) (*bxapiResp.CustomerExploreTopicDetail, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}

	topicService := &engage.ExploreTopicServices{}
	detail, err := topicService.GetTopicDetail(code)
	if err != nil {
		return nil, err
	}

	// Convert featured packages to bxapi response format
	featured := make([]bxapiResp.CustomerExploreTripCard, len(detail.FeaturedPackages))
	for i, p := range detail.FeaturedPackages {
		featured[i] = p
	}

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

	topicService := &engage.ExploreTopicServices{}
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
	for i, p := range result.Packages {
		packages[i] = p
	}

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
		Packages:  packages,
		Total:     result.Total,
		Page:      result.Page,
		PageSize:  result.PageSize,
	}, nil
}

// ============================================================================
// Agent-facing Methods
// ============================================================================

// AgentGetMyTopics returns topics with agent's package counts
func (s *ExploreServices) AgentGetMyTopics(agentID uint) ([]bxapiResp.CustomerExploreTopicCard, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}

	topicService := &engage.ExploreTopicServices{}
	topics, err := topicService.AgentGetMyTopics(agentID)
	if err != nil {
		return nil, err
	}

	// Convert to bxapi response format (using same card structure)
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
			PackageCount: t.PackageCount,
		}
	}

	return cards, nil
}

// AgentAddPackageToTopics adds a package to topics
func (s *ExploreServices) AgentAddPackageToTopics(agentID, packageID uint, topicIDs []uint) error {
	if global.GVA_DB == nil {
		return errors.New("db not init")
	}

	topicService := &engage.ExploreTopicServices{}
	return topicService.AgentAddPackageToTopic(agentID, packageID, topicIDs)
}

// AgentRemovePackageFromTopic removes a package from a topic
func (s *ExploreServices) AgentRemovePackageFromTopic(agentID, packageID, topicID uint) error {
	if global.GVA_DB == nil {
		return errors.New("db not init")
	}

	topicService := &engage.ExploreTopicServices{}
	return topicService.AgentRemovePackageFromTopic(agentID, packageID, topicID)
}

// AgentGetPackageTopics returns topic IDs for a package
func (s *ExploreServices) AgentGetPackageTopics(agentID, packageID uint) ([]uint, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}

	topicService := &engage.ExploreTopicServices{}
	return topicService.AgentGetPackageTopics(agentID, packageID)
}