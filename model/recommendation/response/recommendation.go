package response

// PublicRecommendation the trimmed, customer-facing shape of a recommendation. Internal fields
// (meta, runId, reviewedBy, priority, source) are deliberately not exposed.
type PublicRecommendation struct {
	TargetType string  `json:"targetType"`
	TargetID   uint    `json:"targetId"`
	Placement  string  `json:"placement"`
	Audience   string  `json:"audience"`
	Score      float64 `json:"score"`
	Reason     string  `json:"reason"`
	ReasonAr   string  `json:"reasonAr"`
	IsPinned   bool    `json:"isPinned"`
}

// PublicRecommendationList the public list envelope.
type PublicRecommendationList struct {
	Items    []PublicRecommendation `json:"items"`
	Total    int64                  `json:"total"`
	Page     int                    `json:"page"`
	PageSize int                    `json:"pageSize"`
}

// RecommendationStat per-recommendation engagement numbers for the admin stats view.
type RecommendationStat struct {
	RecommendationID uint    `json:"recommendationId"`
	TargetType       string  `json:"targetType"`
	TargetID         uint    `json:"targetId"`
	Placement        string  `json:"placement"`
	Impressions      int64   `json:"impressions"`
	Clicks           int64   `json:"clicks"`
	Bookings         int64   `json:"bookings"`
	CTR              float64 `json:"ctr"`
}
