package booking

import (
	"time"

	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/model/common"
)

// CustomTripRequest a user's custom-trip builder request (CT- prefixed) that becomes a booking
// once converted/accepted. Single-country for MVP; required documents summarized as JSON.
type CustomTripRequest struct {
	ID                       uint           `json:"id" gorm:"primarykey;comment:ID"`
	RequestCode              string         `json:"requestCode" form:"requestCode" gorm:"size:64;uniqueIndex;comment:CT- prefixed code"`
	UserID                   uint           `json:"userId" form:"userId" gorm:"index;comment:Requesting user ID"`
	AgentID                  uint           `json:"agentId" form:"agentId" gorm:"index;comment:Assigned agent ID"`
	DestinationCountryID     uint           `json:"destinationCountryId" form:"destinationCountryId" gorm:"index;comment:Destination country ID"`
	EstimatedCost            int64          `json:"estimatedCost" form:"estimatedCost" gorm:"default:0;comment:Estimated cost, minor units"`
	RequiredDocumentsSummary common.JSONMap `json:"requiredDocumentsSummary" form:"requiredDocumentsSummary" gorm:"comment:Required document summary"`
	TravelersCount           int            `json:"travelersCount" form:"travelersCount" gorm:"default:1;comment:Number of travelers"`
	Status                   string         `json:"status" form:"status" gorm:"size:32;index;default:pending;comment:Pending/status"`
	SubmittedAt              *time.Time     `json:"submittedAt" form:"submittedAt" gorm:"comment:Submitted at"`
	global.GVA_MODEL
}

func (CustomTripRequest) TableName() string { return "custom_trip_requests" }
