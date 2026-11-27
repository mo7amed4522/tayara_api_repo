package recommendation

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/mo7amed4522/tayara_api/constants"
	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/model/common"
	"github.com/mo7amed4522/tayara_api/model/recommendation"
	"github.com/mo7amed4522/tayara_api/model/recommendation/request"
	"gorm.io/gorm"
)

// @author:		[mo7amed4522](https://github.com/mo7amed4522)
// @function:		GetSetting
// @description:	Read the engine setting for a placement (defaults when the row is absent)
// @param:			placement string
// @return:		setting *recommendation.RecommendationSetting, err error
func (s *RecommendationServices) GetSetting(placement string) (setting *recommendation.RecommendationSetting, err error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	if strings.TrimSpace(placement) == "" {
		return nil, errors.New("placement is required")
	}
	var row recommendation.RecommendationSetting
	if err := global.GVA_DB.Where("placement = ?", placement).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return defaultSetting(placement), nil
		}
		return nil, err
	}
	if row.MaxItems <= 0 {
		row.MaxItems = defaultMaxItems
	}
	if row.CandidateLimit <= 0 {
		row.CandidateLimit = 50
	}
	if row.AutoApproveThreshold < 0 || row.AutoApproveThreshold > 1 {
		row.AutoApproveThreshold = 0.8
	}
	return &row, nil
}

// findSetting is the engine's read path; it never creates a row as a side effect.
func (s *RecommendationServices) findSetting(placement string) (*recommendation.RecommendationSetting, error) {
	return s.GetSetting(placement)
}

// @author:		[mo7amed4522](https://github.com/mo7amed4522)
// @function:		UpdateSetting
// @description:	Upsert the engine setting for a placement from an explicit field map
// @param:			req request.UpdateSettingReq
// @return:		setting *recommendation.RecommendationSetting, err error
func (s *RecommendationServices) UpdateSetting(req request.UpdateSettingReq) (setting *recommendation.RecommendationSetting, err error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	placement := strings.TrimSpace(req.Placement)
	if placement == "" {
		return nil, errors.New("placement is required")
	}

	updates := map[string]interface{}{}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}
	if req.MaxItems != nil {
		if *req.MaxItems <= 0 {
			return nil, errors.New("maxItems must be positive")
		}
		updates["max_items"] = *req.MaxItems
	}
	if req.AutoApprove != nil {
		updates["auto_approve"] = *req.AutoApprove
	}
	if req.AutoApproveThreshold != nil {
		if *req.AutoApproveThreshold < 0 || *req.AutoApproveThreshold > 1 {
			return nil, errors.New("autoApproveThreshold must be between 0 and 1")
		}
		updates["auto_approve_threshold"] = *req.AutoApproveThreshold
	}
	if req.CandidateLimit != nil {
		if *req.CandidateLimit <= 0 {
			return nil, errors.New("candidateLimit must be positive")
		}
		updates["candidate_limit"] = *req.CandidateLimit
	}
	if req.TargetTypes != nil {
		raw, err := json.Marshal(req.TargetTypes)
		if err != nil {
			return nil, err
		}
		types := common.JSONMap{}
		if err := json.Unmarshal(raw, &types); err != nil {
			return nil, fmt.Errorf("targetTypes must be a JSON object: %w", err)
		}
		updates["target_types"] = types
	}
	if req.PromptTemplate != nil {
		updates["prompt_template"] = *req.PromptTemplate
	}
	if req.PromptVersion != nil {
		updates["prompt_version"] = *req.PromptVersion
	}
	if req.RefreshCron != nil {
		updates["refresh_cron"] = *req.RefreshCron
	}
	if req.ExpireAfterHours != nil {
		updates["expire_after_hours"] = *req.ExpireAfterHours
	}

	var row recommendation.RecommendationSetting
	err = global.GVA_DB.Where("placement = ?", placement).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		row = *defaultSetting(placement)
		row.Placement = placement
		if err := global.GVA_DB.Create(&row).Error; err != nil {
			return nil, err
		}
		if len(updates) == 0 {
			return &row, nil
		}
		if err := global.GVA_DB.Model(&row).Updates(updates).Error; err != nil {
			return nil, err
		}
		return s.GetSetting(placement)
	}
	if err != nil {
		return nil, err
	}
	if len(updates) == 0 {
		return &row, nil
	}
	if err := global.GVA_DB.Model(&row).Updates(updates).Error; err != nil {
		return nil, err
	}
	return s.GetSetting(placement)
}

// defaultSetting the in-code defaults, mirroring the seeded rows in source/recommendation.
func defaultSetting(placement string) *recommendation.RecommendationSetting {
	return &recommendation.RecommendationSetting{
		Placement:            placement,
		Enabled:              false,
		MaxItems:             defaultMaxItems,
		AutoApprove:          false,
		AutoApproveThreshold: 0.8,
		CandidateLimit:       50,
		TargetTypes:          common.JSONMap{"targetTypes": []string{constants.RecommendationTargetTypePackage, constants.RecommendationTargetTypeTrip, constants.RecommendationTargetTypeAgency}},
		PromptTemplate:       DefaultPromptTemplate,
		PromptVersion:        "v1",
		RefreshCron:          "6h",
		ExpireAfterHours:     24,
	}
}
