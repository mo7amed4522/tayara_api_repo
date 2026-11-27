package recommendation

import (
	"time"

	"github.com/mo7amed4522/tayara_api/global"
)

// RecommendationRun an audit record of one recommendation generation attempt: the model, the
// prompt version, the candidate snapshot, the raw output, and how it ended.
type RecommendationRun struct {
	ID             uint       `json:"id" gorm:"primarykey;comment:id of run"`
	Trigger        string     `json:"trigger" form:"trigger" gorm:"size:16;index;comment:cron | manual"`
	TriggeredBy    uint       `json:"triggeredBy" form:"triggeredBy" gorm:"index;comment:Admin who triggered it (0 for cron)"`
	Placement      string     `json:"placement" form:"placement" gorm:"size:64;index;comment:Placement the run targets"`
	Audience       string     `json:"audience" form:"audience" gorm:"size:64;index;comment:Audience segment"`
	Provider       string     `json:"provider" form:"provider" gorm:"size:32;comment:LLM provider"`
	ModelName      string     `json:"modelName" form:"modelName" gorm:"size:64;comment:LLM model name"`
	PromptVersion  string     `json:"promptVersion" form:"promptVersion" gorm:"size:16;comment:Prompt template version"`
	InputSnapshot  string     `json:"inputSnapshot" form:"inputSnapshot" gorm:"type:longtext;comment:Candidate snapshot sent to the LLM (admin only)"`
	RawOutput      string     `json:"rawOutput" form:"rawOutput" gorm:"type:longtext;comment:Raw LLM output"`
	Status         string     `json:"status" form:"status" gorm:"size:16;index;default:running;comment:running | success | failed | fallback"`
	ErrorMessage   string     `json:"errorMessage" form:"errorMessage" gorm:"type:text;comment:Failure reason when the run fell back"`
	CandidateCount int        `json:"candidateCount" form:"candidateCount" gorm:"comment:Candidates offered to the LLM"`
	ResultCount    int        `json:"resultCount" form:"resultCount" gorm:"comment:Items persisted from the run"`
	DurationMs     int64      `json:"durationMs" form:"durationMs" gorm:"comment:Wall-clock duration in milliseconds"`
	StartedAt      *time.Time `json:"startedAt" form:"startedAt" gorm:"comment:Run start"`
	FinishedAt     *time.Time `json:"finishedAt" form:"finishedAt" gorm:"comment:Run finish"`

	global.GVA_MODEL
}

func (RecommendationRun) TableName() string { return "recommendation_run" }
