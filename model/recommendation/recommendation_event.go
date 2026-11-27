package recommendation

import (
	"github.com/mo7amed4522/tayara_api/global"
)

// RecommendationEvent a per-impression/click/booking signal for a stored recommendation.
// UserID is nullable and stores the numeric ID only (no PII); signals are only ever used aggregated.
type RecommendationEvent struct {
	ID               uint   `json:"id" gorm:"primarykey;comment:id of event"`
	RecommendationID uint   `json:"recommendationId" form:"recommendationId" gorm:"index;comment:Recommendation row the event belongs to"`
	EventType        string `json:"eventType" form:"eventType" gorm:"size:16;index;comment:impression | click | booking"`
	UserID           *uint  `json:"userId" form:"userId" gorm:"index;comment:Acting user ID (nullable, no PII)"`
	Placement        string `json:"placement" form:"placement" gorm:"size:64;index;comment:Placement shown on"`

	global.GVA_MODEL
}

func (RecommendationEvent) TableName() string { return "recommendation_event" }
