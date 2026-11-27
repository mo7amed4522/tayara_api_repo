package recommendation

// RecommendationCandidate the only shape the LLM ever sees per item: public catalog fields and
// aggregated signals. No names, emails, phones or any per-user data are included.
type RecommendationCandidate struct {
	TargetType   string  `json:"targetType"`
	TargetID     uint    `json:"targetId"`
	Title        string  `json:"title"`
	Category     string  `json:"category,omitempty"`
	Destination  string  `json:"destination,omitempty"`
	PriceBand    string  `json:"priceBand,omitempty"`
	Rating       float64 `json:"rating,omitempty"`
	ReviewCount  int64   `json:"reviewCount"`
	Bookings30d  int64   `json:"bookings30d"`
	RecencyDays  int     `json:"recencyDays"`
	AgencyRating float64 `json:"agencyRating,omitempty"`
	Impressions  int64   `json:"impressions,omitempty"`
	Clicks       int64   `json:"clicks,omitempty"`
	Bookings     int64   `json:"bookings,omitempty"`
}

// RecommendationItem one validated recommendation returned by the LLM (or produced by the rules
// fallback) before it is persisted.
type RecommendationItem struct {
	TargetType string  `json:"targetType"`
	TargetID   uint    `json:"targetId"`
	Score      float64 `json:"score"`
	Reason     string  `json:"reason"`
}

// RecommendationSignal aggregated engagement counts per target, attached to candidates as context.
type RecommendationSignal struct {
	TargetType  string `gorm:"column:target_type"`
	TargetID    uint   `gorm:"column:target_id"`
	Impressions int64  `gorm:"column:impressions"`
	Clicks      int64  `gorm:"column:clicks"`
	Bookings    int64  `gorm:"column:bookings"`
}
