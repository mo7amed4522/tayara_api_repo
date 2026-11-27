package request

import (
	"time"

	"github.com/mo7amed4522/tayara_api/model/common/request"
)

// RecommendationSearch admin list filters for stored recommendations.
type RecommendationSearch struct {
	request.PageInfo
	TargetType string `json:"targetType" form:"targetType"`
	Placement  string `json:"placement" form:"placement"`
	Audience   string `json:"audience" form:"audience"`
	Status     string `json:"status" form:"status"`
	Source     string `json:"source" form:"source"`
	Keyword    string `json:"keyword" form:"keyword"`
	StartTime  string `json:"startTime" form:"startTime"` // YYYY-MM-DD or RFC3339
	EndTime    string `json:"endTime" form:"endTime"`
}

// BaseTimeRange narrows a stats query.
type BaseTimeRange struct {
	Placement string `json:"placement" form:"placement"`
	StartTime string `json:"startTime" form:"startTime"`
	EndTime   string `json:"endTime" form:"endTime"`
}

// CreateRecommendationReq a manual admin entry (source = admin).
type CreateRecommendationReq struct {
	TargetType string     `json:"targetType" form:"targetType" binding:"required,oneof=package trip agency"`
	TargetID   uint       `json:"targetId" form:"targetId" binding:"required"`
	Placement  string     `json:"placement" form:"placement" binding:"required"`
	Audience   string     `json:"audience" form:"audience"`
	Score      float64    `json:"score" form:"score"`
	Reason     string     `json:"reason" form:"reason"`
	ReasonAr   string     `json:"reasonAr" form:"reasonAr"`
	Priority   int        `json:"priority" form:"priority"`
	IsPinned   bool       `json:"isPinned" form:"isPinned"`
	StartsAt   *time.Time `json:"startsAt" form:"startsAt"`
	EndsAt     *time.Time `json:"endsAt" form:"endsAt"`
}

// UpdateRecommendationReq the only fields an admin may edit on a stored recommendation.
type UpdateRecommendationReq struct {
	ID        uint       `json:"id" form:"id" binding:"required"`
	Reason    *string    `json:"reason" form:"reason"`
	ReasonAr  *string    `json:"reasonAr" form:"reasonAr"`
	Priority  *int       `json:"priority" form:"priority"`
	IsPinned  *bool      `json:"isPinned" form:"isPinned"`
	StartsAt  *time.Time `json:"startsAt" form:"startsAt"`
	EndsAt    *time.Time `json:"endsAt" form:"endsAt"`
	Placement string     `json:"placement" form:"placement"`
	Audience  string     `json:"audience" form:"audience"`
}

// ReviewRecommendationReq one admin decision on one row.
type ReviewRecommendationReq struct {
	ID     uint   `json:"id" form:"id" binding:"required"`
	Action string `json:"action" form:"action" binding:"required,oneof=approve reject"`
}

// BatchReviewReq the same decision applied to several rows in one transaction.
type BatchReviewReq struct {
	IDs    []uint `json:"ids" form:"ids" binding:"required"`
	Action string `json:"action" form:"action" binding:"required,oneof=approve reject"`
}

// GenerateReq triggers an admin LLM run for one placement/audience.
type GenerateReq struct {
	Placement string `json:"placement" form:"placement" binding:"required"`
	Audience  string `json:"audience" form:"audience"`
}

// RunSearch admin audit list filters for generation runs.
type RunSearch struct {
	request.PageInfo
	Placement string `json:"placement" form:"placement"`
	Status    string `json:"status" form:"status"`
	Trigger   string `json:"trigger" form:"trigger"`
}

// UpdateSettingReq admin edit of a placement's engine settings.
type UpdateSettingReq struct {
	Placement            string      `json:"placement" form:"placement" binding:"required"`
	Enabled              *bool       `json:"enabled" form:"enabled"`
	MaxItems             *int        `json:"maxItems" form:"maxItems"`
	AutoApprove          *bool       `json:"autoApprove" form:"autoApprove"`
	AutoApproveThreshold *float64    `json:"autoApproveThreshold" form:"autoApproveThreshold"`
	CandidateLimit       *int        `json:"candidateLimit" form:"candidateLimit"`
	TargetTypes          interface{} `json:"targetTypes" form:"targetTypes"`
	PromptTemplate       *string     `json:"promptTemplate" form:"promptTemplate"`
	PromptVersion        *string     `json:"promptVersion" form:"promptVersion"`
	RefreshCron          *string     `json:"refreshCron" form:"refreshCron"`
	ExpireAfterHours     *int        `json:"expireAfterHours" form:"expireAfterHours"`
}

// RecommendationEventReq a public impression/click/booking signal.
type RecommendationEventReq struct {
	RecommendationID uint   `json:"recommendationId" form:"recommendationId" binding:"required"`
	EventType        string `json:"eventType" form:"eventType" binding:"required,oneof=impression click booking"`
	Placement        string `json:"placement" form:"placement"`
}
