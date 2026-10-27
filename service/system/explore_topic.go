package system

import (
	"errors"

	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/model/engage/request"
	"github.com/mo7amed4522/tayara_api/model/engage/response"
	"github.com/mo7amed4522/tayara_api/service/engage"
	"github.com/mo7amed4522/tayara_api/utils"
)

// ExploreTopicServices provides admin endpoints for explore topics
type ExploreTopicServices struct{}

// AdminListTopics returns paginated list of all topics
// GET /admin/explore/topics
func (s *ExploreTopicServices) AdminListTopics(req request.AdminExploreTopicPageReq) (*utils.PaginateResult, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}

	topicService := &engage.ExploreTopicServices{}
	return topicService.AdminListTopics(req)
}

// AdminGetTopic returns a single topic by ID
// GET /admin/explore/topics/:id
func (s *ExploreTopicServices) AdminGetTopic(id uint) (*response.AdminExploreTopicDetail, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}

	topicService := &engage.ExploreTopicServices{}
	topic, err := topicService.AdminGetTopic(id)
	if err != nil {
		return nil, err
	}

	// Get package count
	counts := topicService.GetPackageCountsByTopic([]uint{id})

	return &response.AdminExploreTopicDetail{
		ID:           topic.ID,
		Code:         topic.Code,
		Name:         topic.Name,
		NameAr:       topic.NameAr,
		Description:  topic.Description,
		ImageURL:     topic.ImageURL,
		IconURL:      topic.IconURL,
		SortOrder:    topic.SortOrder,
		IsActive:     topic.IsActive,
		PackageCount: counts[id],
		CreatedAt:    topic.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:    topic.UpdatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

// AdminCreateTopic creates a new explore topic
// POST /admin/explore/topics
func (s *ExploreTopicServices) AdminCreateTopic(req request.ExploreTopicReq) (*response.AdminExploreTopicDetail, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}

	topicService := &engage.ExploreTopicServices{}
	topic, err := topicService.AdminCreateTopic(req)
	if err != nil {
		return nil, err
	}

	return &response.AdminExploreTopicDetail{
		ID:           topic.ID,
		Code:         topic.Code,
		Name:         topic.Name,
		NameAr:       topic.NameAr,
		Description:  topic.Description,
		ImageURL:     topic.ImageURL,
		IconURL:      topic.IconURL,
		SortOrder:    topic.SortOrder,
		IsActive:     topic.IsActive,
		PackageCount: 0,
		CreatedAt:    topic.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:    topic.UpdatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

// AdminUpdateTopic updates an existing topic
// PUT /admin/explore/topics/:id
func (s *ExploreTopicServices) AdminUpdateTopic(id uint, req request.ExploreTopicReq) (*response.AdminExploreTopicDetail, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}

	topicService := &engage.ExploreTopicServices{}
	topic, err := topicService.AdminUpdateTopic(id, req)
	if err != nil {
		return nil, err
	}

	counts := topicService.GetPackageCountsByTopic([]uint{id})

	return &response.AdminExploreTopicDetail{
		ID:           topic.ID,
		Code:         topic.Code,
		Name:         topic.Name,
		NameAr:       topic.NameAr,
		Description:  topic.Description,
		ImageURL:     topic.ImageURL,
		IconURL:      topic.IconURL,
		SortOrder:    topic.SortOrder,
		IsActive:     topic.IsActive,
		PackageCount: counts[id],
		CreatedAt:    topic.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:    topic.UpdatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

// AdminDeleteTopic deletes a topic
// DELETE /admin/explore/topics/:id
func (s *ExploreTopicServices) AdminDeleteTopic(id uint) error {
	if global.GVA_DB == nil {
		return errors.New("db not init")
	}

	topicService := &engage.ExploreTopicServices{}
	return topicService.AdminDeleteTopic(id)
}

// AdminReorderTopics bulk updates topic sort order
// PUT /admin/explore/topics/reorder
func (s *ExploreTopicServices) AdminReorderTopics(req []request.TopicOrderReq) error {
	if global.GVA_DB == nil {
		return errors.New("db not init")
	}

	topicService := &engage.ExploreTopicServices{}
	return topicService.AdminReorderTopics(req)
}

// AdminManageTopicPackages bulk manages packages in a topic
// POST /admin/explore/topics/:id/packages
func (s *ExploreTopicServices) AdminManageTopicPackages(topicID uint, req request.AdminManageTopicPackagesReq) error {
	if global.GVA_DB == nil {
		return errors.New("db not init")
	}

	topicService := &engage.ExploreTopicServices{}
	return topicService.AdminManageTopicPackages(topicID, req.PackageIDs, req.Replace)
}

// AdminGetTopicPackages returns packages associated with a topic
// GET /admin/explore/topics/:id/packages
func (s *ExploreTopicServices) AdminGetTopicPackages(topicID uint, page, pageSize int) (*utils.PaginateResult, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}

	topicService := &engage.ExploreTopicServices{}
	return topicService.AdminGetTopicPackages(topicID, page, pageSize)
}