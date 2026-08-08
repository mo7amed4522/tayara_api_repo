package engage

import (
	"errors"
	"sort"
	"time"

	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/model/engage"
	engageResp "github.com/mo7amed4522/tayara_api/model/engage/response"
)

// AdGroupedAnalytics aggregates ad impression/CTR grouped by placement within an optional
// created-at window [from, to] (nil bound = every time). Counters are the cumulative
// AdCreative counters; the window only bounds which creatives are included. Admin-scoped read.
func (a AdServices) AdGroupedAnalytics(from, to *time.Time) (*engageResp.AdAnalytics, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	db := global.GVA_DB.Model(&engage.AdCreative{})
	if from != nil {
		db = db.Where("created_at >= ?", *from)
	}
	if to != nil {
		db = db.Where("created_at <= ?", *to)
	}
	var ads []engage.AdCreative
	if err := db.Find(&ads).Error; err != nil {
		return nil, err
	}

	out := &engageResp.AdAnalytics{}
	byPlc := map[string]*engageResp.AdGroupedRow{}
	for i := range ads {
		ad := ads[i]
		bucket, ok := byPlc[ad.Placement]
		if !ok {
			bucket = &engageResp.AdGroupedRow{Placement: ad.Placement}
			byPlc[ad.Placement] = bucket
		}
		bucket.Ads++
		bucket.Impressions += ad.Impressions
		bucket.Clicks += ad.Clicks
		out.TotalAds++
		out.TotalImpressions += ad.Impressions
		out.TotalClicks += ad.Clicks
	}
	for placement, bucket := range byPlc {
		if bucket.Impressions > 0 {
			bucket.CTR = float64(bucket.Clicks) / float64(bucket.Impressions) * 100
		}
		bucket.Placement = placement
		out.ByPlacement = append(out.ByPlacement, *bucket)
	}
	if out.TotalImpressions > 0 {
		out.OverallCTR = float64(out.TotalClicks) / float64(out.TotalImpressions) * 100
	}

	// top 10 creatives by impressions
	sort.Slice(ads, func(i, j int) bool { return ads[i].Impressions > ads[j].Impressions })
	for i := 0; i < len(ads) && len(out.TopAds) < 10; i++ {
		ctr := float64(0)
		if ads[i].Impressions > 0 {
			ctr = float64(ads[i].Clicks) / float64(ads[i].Impressions) * 100
		}
		out.TopAds = append(out.TopAds, engageResp.AdDetail{AdCreative: ads[i], CTR: ctr})
	}
	return out, nil
}
