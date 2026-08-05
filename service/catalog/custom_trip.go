package catalog

import (
	"errors"

	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/model/catalog"
	"github.com/mo7amed4522/tayara_api/model/catalog/request"
	"gorm.io/gorm"
)

type CustomTripServices struct{}

// CreateTemplate creates a custom-trip builder configuration owned by the agent, plus its countries.
func (c *CustomTripServices) CreateTemplate(agentID uint, req request.CustomTripTemplateReq) (*catalog.CustomTripTemplate, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	tpl := &catalog.CustomTripTemplate{
		AgentID:            agentID,
		Name:               req.Name,
		DurationMinDays:    req.DurationMinDays,
		DurationMaxDays:    req.DurationMaxDays,
		TransportTypes:     req.TransportTypes,
		HotelLevels:        req.HotelLevels,
		Activities:         req.Activities,
		VisaAddonsEnabled:  req.VisaAddonsEnabled,
		PricingLogicHidden: req.PricingLogicHidden,
		IsActive:           req.IsActive,
	}
	err := global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(tpl).Error; err != nil {
			return err
		}
		return replaceTemplateCountries(tx, tpl.ID, req.DestinationCountryIDs)
	})
	return tpl, err
}

// UpdateTemplate updates one of the agent's own templates and replaces its countries.
func (c *CustomTripServices) UpdateTemplate(agentID uint, req request.CustomTripTemplateReq) (*catalog.CustomTripTemplate, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	var tpl catalog.CustomTripTemplate
	if err := global.GVA_DB.Where("id = ? AND agent_id = ?", req.ID, agentID).First(&tpl).Error; err != nil {
		return nil, errors.New("template not found")
	}
	err := global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&catalog.CustomTripTemplate{}).Where("id = ?", req.ID).Updates(map[string]interface{}{
			"name": req.Name, "duration_min_days": req.DurationMinDays, "duration_max_days": req.DurationMaxDays,
			"transport_types": req.TransportTypes, "hotel_levels": req.HotelLevels, "activities": req.Activities,
			"visa_addons_enabled": req.VisaAddonsEnabled, "pricing_logic_hidden": req.PricingLogicHidden, "is_active": req.IsActive,
		}).Error; err != nil {
			return err
		}
		if err := tx.Where("template_id = ?", req.ID).Delete(&catalog.CustomTripTemplateCountry{}).Error; err != nil {
			return err
		}
		return replaceTemplateCountries(tx, req.ID, req.DestinationCountryIDs)
	})
	if err != nil {
		return nil, err
	}
	return c.GetOwnTemplate(agentID, req.ID)
}

// ListOwnTemplates returns the agent's templates (agent scoping enforced).
func (c *CustomTripServices) ListOwnTemplates(agentID uint) ([]catalog.CustomTripTemplate, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	var list []catalog.CustomTripTemplate
	if err := global.GVA_DB.Where("agent_id = ?", agentID).Order("updated_at desc").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// GetOwnTemplate returns one of the agent's own templates.
func (c *CustomTripServices) GetOwnTemplate(agentID, id uint) (*catalog.CustomTripTemplate, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	var tpl catalog.CustomTripTemplate
	if err := global.GVA_DB.Where("id = ? AND agent_id = ?", id, agentID).First(&tpl).Error; err != nil {
		return nil, errors.New("template not found")
	}
	return &tpl, nil
}

// DeleteTemplate soft-deletes one of the agent's own templates.
func (c *CustomTripServices) DeleteTemplate(agentID, id uint) error {
	if global.GVA_DB == nil {
		return errors.New("db not init")
	}
	return global.GVA_DB.Where("id = ? AND agent_id = ?", id, agentID).Delete(&catalog.CustomTripTemplate{}).Error
}

// replaceTemplateCountries persists the countries for a template (called inside a transaction).
func replaceTemplateCountries(tx *gorm.DB, templateID uint, countryIDs []uint) error {
	if len(countryIDs) == 0 {
		return nil
	}
	rows := make([]catalog.CustomTripTemplateCountry, 0, len(countryIDs))
	for _, cid := range countryIDs {
		rows = append(rows, catalog.CustomTripTemplateCountry{TemplateID: templateID, CountryID: cid})
	}
	return tx.Create(&rows).Error
}
