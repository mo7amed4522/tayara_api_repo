package response

import (
	"github.com/mo7amed4522/tayara_api/model/engage"
)

// AdDetail an ad creative plus its click-through-rate (computed).
type AdDetail struct {
	engage.AdCreative
	CTR float64 `json:"ctr"` // clicks / impressions * 100
}
