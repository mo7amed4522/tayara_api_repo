package recommendation

import (
	"time"

	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/model/common"
)

// Recommendation a stored, admin-reviewable recommendation of a catalog item
// (package / trip / agency) for a placement + audience. Only approved rows are served to customers.
type Recommendation struct {
	ID         uint           `json:"id" gorm:"primarykey;comment:id of recommendation"`
	TargetType string         `json:"targetType" form:"targetType" gorm:"size:32;uniqueIndex:idx_rec_target_placement_audience;index:idx_rec_target;comment:package | trip | agency"`
	TargetID   uint           `json:"targetId" form:"targetId" gorm:"uniqueIndex:idx_rec_target_placement_audience;index:idx_rec_target;comment:ID of the recommended item"`
	Placement  string         `json:"placement" form:"placement" gorm:"size:64;uniqueIndex:idx_rec_target_placement_audience;index;comment:Where it is shown e.g. home_top, search, package_detail"`
	Audience   string         `json:"audience" form:"audience" gorm:"size:64;uniqueIndex:idx_rec_target_placement_audience;index;comment:Audience segment e.g. all, new_user, family"`
	Source     string         `json:"source" form:"source" gorm:"size:16;comment:llm | rules | admin"`
	Status     string         `json:"status" form:"status" gorm:"size:16;index;default:pending;comment:pending | approved | rejected | expired"`
	Score      float64        `json:"score" form:"score" gorm:"comment:Relevance score 0-1"`
	Reason     string         `json:"reason" form:"reason" gorm:"type:text;comment:Short human-readable reason"`
	ReasonAr   string         `json:"reasonAr" form:"reasonAr" gorm:"type:text;comment:Arabic reason"`
	Priority   int            `json:"priority" form:"priority" gorm:"default:0;comment:Admin ordering, higher first"`
	IsPinned   bool           `json:"isPinned" form:"isPinned" gorm:"default:false;comment:Pinned by admin, never replaced by a run"`
	StartsAt   *time.Time     `json:"startsAt" form:"startsAt" gorm:"comment:Start of visibility"`
	EndsAt     *time.Time     `json:"endsAt" form:"endsAt" gorm:"comment:End of visibility"`
	RunID      uint           `json:"runId" form:"runId" gorm:"index;comment:RecommendationRun that created it"`
	ReviewedBy uint           `json:"reviewedBy" form:"reviewedBy" gorm:"comment:Admin who reviewed"`
	ReviewedAt *time.Time     `json:"reviewedAt" form:"reviewedAt" gorm:"comment:Review time"`
	Meta       common.JSONMap `json:"meta" form:"meta" gorm:"type:json;comment:Extra data"`

	global.GVA_MODEL
}

func (Recommendation) TableName() string { return "recommendation" }
