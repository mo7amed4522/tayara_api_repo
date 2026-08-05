package request

import "github.com/mo7amed4522/tayara_api/model/common"

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
