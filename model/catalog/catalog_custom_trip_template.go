package catalog

import (
	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/model/common"
)

// CustomTripTemplate an agent's configuration for the custom-trip builder (single-country MVP).
// Repeated structured lists (transport, hotels, activities) are stored as JSONMap arrays.
type CustomTripTemplate struct {
	ID                 uint           `json:"id" gorm:"primarykey;comment:ID"`
	AgentID            uint           `json:"agentId" form:"agentId" gorm:"index;comment:Owning agent ID"`
	Name               string         `json:"name" form:"name" gorm:"size:191;index;comment:Template name"`
	DurationMinDays    int            `json:"durationMinDays" form:"durationMinDays" gorm:"default:1;comment:Min duration days"`
	DurationMaxDays    int            `json:"durationMaxDays" form:"durationMaxDays" gorm:"default:30;comment:Max duration days"`
	TransportTypes     common.JSONMap `json:"transportTypes" form:"transportTypes" gorm:"comment:Transport options"`
	HotelLevels        common.JSONMap `json:"hotelLevels" form:"hotelLevels" gorm:"comment:Hotel star levels"`
	Activities         common.JSONMap `json:"activities" form:"activities" gorm:"comment:Predefined activities"`
	VisaAddonsEnabled  bool           `json:"visaAddonsEnabled" form:"visaAddonsEnabled" gorm:"default:false;comment:Visa add-on enabled"`
	PricingLogicHidden bool           `json:"pricingLogicHidden" form:"pricingLogicHidden" gorm:"default:true;comment:Only estimate shown to user"`
	IsActive           bool           `json:"isActive" form:"isActive" gorm:"default:true;comment:Is active"`
	global.GVA_MODEL
}

func (CustomTripTemplate) TableName() string { return "catalog_custom_trip_template" }
