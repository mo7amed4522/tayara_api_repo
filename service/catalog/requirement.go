package catalog

import (
	"errors"

	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/model/catalog"
	catReq "github.com/mo7amed4522/tayara_api/model/catalog/request"
	"github.com/mo7amed4522/tayara_api/utils"
	"gorm.io/gorm"
)

// RequirementServices owns the custom-trip requirement profiles (per template+country documents).
type RequirementServices struct{}

// SetProfiles replaces the requirement profiles for a template+country (agent-owned template).
func (s RequirementServices) SetProfiles(agentID uint, req catReq.SetRequirementProfilesReq) error {
	if global.GVA_DB == nil {
		return errors.New("db not init")
	}
	var tpl catalog.CustomTripTemplate
	if err := global.GVA_DB.Where("id = ? AND agent_id = ?", req.TemplateID, agentID).First(&tpl).Error; err != nil {
		return errors.New("template not found")
	}
	return global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("template_id = ? AND country_id = ?", req.TemplateID, req.CountryID).
			Delete(&catalog.CustomTripRequirementProfile{}).Error; err != nil {
			return err
		}
		if len(req.Profiles) == 0 {
			return nil
		}
		rows := make([]catalog.CustomTripRequirementProfile, 0, len(req.Profiles))
		for _, p := range req.Profiles {
			phase := p.Phase
			if phase == "" {
				phase = "before"
			}
			rows = append(rows, catalog.CustomTripRequirementProfile{
				TemplateID: req.TemplateID, CountryID: req.CountryID,
				DocumentTypeID: p.DocumentTypeID, TravelerField: p.TravelerField,
				IsRequired: p.IsRequired, Phase: phase, Notes: p.Notes,
			})
		}
		return tx.Create(&rows).Error
	})
}

// GetProfiles returns the requirement profiles for a template+country (agent-owned).
func (s RequirementServices) GetProfiles(agentID, templateID, countryID uint, page, pageSize int) ([]catalog.CustomTripRequirementProfile, int64, int64, error) {
	if global.GVA_DB == nil {
		return nil, 0, 0, errors.New("db not init")
	}
	var tpl catalog.CustomTripTemplate
	if err := global.GVA_DB.Where("id = ? AND agent_id = ?", templateID, agentID).First(&tpl).Error; err != nil {
		return nil, 0, 0, errors.New("template not found")
	}
	var list []catalog.CustomTripRequirementProfile
	countDB := global.GVA_DB.Model(&catalog.CustomTripRequirementProfile{}).
		Where("template_id = ? AND country_id = ?", templateID, countryID)
	queryDB := global.GVA_DB.Model(&catalog.CustomTripRequirementProfile{}).
		Where("template_id = ? AND country_id = ?", templateID, countryID).
		Order("id asc")
	pr, err := utils.PaginateWithCustomQuery(countDB, queryDB, page, pageSize, &list)
	if err != nil {
		return nil, 0, 0, err
	}
	return list, pr.Total, pr.TotalPage, nil
}
