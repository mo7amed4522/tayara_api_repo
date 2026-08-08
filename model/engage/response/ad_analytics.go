package response

// AdGroupedRow is one bucket of a grouped ad-analytics breakdown (e.g. per placement).
type AdGroupedRow struct {
	Placement   string  `json:"placement"`
	Ads         int64   `json:"ads"`
	Impressions int64   `json:"impressions"`
	Clicks      int64   `json:"clicks"`
	CTR         float64 `json:"ctr"` // clicks / impressions * 100
}

// AdAnalytics is the aggregated view of ad impression/CTR grouped by placement, plus the
// top performing creatives in the range. All counters are the stored cumulative AdCreative
// counters matched by the optional created-at window.
type AdAnalytics struct {
	ByPlacement      []AdGroupedRow `json:"byPlacement"`
	TopAds           []AdDetail     `json:"topAds"` // top 10 by impressions
	TotalAds         int64          `json:"totalAds"`
	TotalImpressions int64          `json:"totalImpressions"`
	TotalClicks      int64          `json:"totalClicks"`
	OverallCTR       float64        `json:"overallCtr"`
}
