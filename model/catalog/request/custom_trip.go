package request

import "github.com/mo7amed4522/tayara_api/model/common"

// RequirementProfileReq one dynamic document requirement for a template+country.
type RequirementProfileReq struct {
	DocumentTypeID uint   `json:"documentTypeId"`
	TravelerField  string `json:"travelerField"`
	IsRequired     bool   `json:"isRequired"`
	Phase          string `json:"phase"` // before|after
	Notes          string `json:"notes"`
}

// SetRequirementProfilesReq replaces a template+country's requirement profiles (agent-owned).
type SetRequirementProfilesReq struct {
	TemplateID uint                    `json:"templateId" binding:"required"`
	CountryID  uint                    `json:"countryId" binding:"required"`
	Profiles   []RequirementProfileReq `json:"profiles"`
}

// CustomTripTemplateReq create/update payload for an agent's custom-trip builder configuration.
type CustomTripTemplateReq struct {
	ID                    uint           `json:"id"` // 0 on create
	Name                  string         `json:"name"`
	DestinationCountryIDs []uint         `json:"destinationCountryIDs"`
	DurationMinDays       int            `json:"durationMinDays"`
	DurationMaxDays       int            `json:"durationMaxDays"`
	TransportTypes        common.JSONMap `json:"transportTypes"`
	HotelLevels           common.JSONMap `json:"hotelLevels"`
	Activities            common.JSONMap `json:"activities"`
	VisaAddonsEnabled     bool           `json:"visaAddonsEnabled"`
	PricingLogicHidden    bool           `json:"pricingLogicHidden"`
	IsActive              bool           `json:"isActive"`
}
