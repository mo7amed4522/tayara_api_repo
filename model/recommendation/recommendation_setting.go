package recommendation

import (
	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/model/common"
)

// RecommendationSetting the per-placement configuration the engine reads: whether generation is
// enabled, how many items to keep, whether LLM output auto-approves, and the prompt template.
type RecommendationSetting struct {
	ID                   uint           `json:"id" gorm:"primarykey;comment:id of setting"`
	Placement            string         `json:"placement" form:"placement" gorm:"size:64;uniqueIndex;comment:Placement key e.g. home_top"`
	Enabled              bool           `json:"enabled" form:"enabled" gorm:"default:false;comment:Generate recommendations for this placement"`
	MaxItems             int            `json:"maxItems" form:"maxItems" gorm:"default:10;comment:Maximum items to keep per run"`
	AutoApprove          bool           `json:"autoApprove" form:"autoApprove" gorm:"default:false;comment:Approve LLM output automatically"`
	AutoApproveThreshold float64        `json:"autoApproveThreshold" form:"autoApproveThreshold" gorm:"default:0.8;comment:Minimum score for auto-approval"`
	CandidateLimit       int            `json:"candidateLimit" form:"candidateLimit" gorm:"default:50;comment:Maximum candidates offered to the LLM"`
	TargetTypes          common.JSONMap `json:"targetTypes" form:"targetTypes" gorm:"type:json;comment:Allowed target types (package, trip, agency)"`
	PromptTemplate       string         `json:"promptTemplate" form:"promptTemplate" gorm:"type:text;comment:System prompt template"`
	PromptVersion        string         `json:"promptVersion" form:"promptVersion" gorm:"size:16;default:v1;comment:Prompt template version"`
	RefreshCron          string         `json:"refreshCron" form:"refreshCron" gorm:"size:64;comment:Refresh interval, e.g. 6h"`
	ExpireAfterHours     int            `json:"expireAfterHours" form:"expireAfterHours" gorm:"default:24;comment:Rows not returned by a run expire after this many hours"`

	global.GVA_MODEL
}

func (RecommendationSetting) TableName() string { return "recommendation_setting" }
