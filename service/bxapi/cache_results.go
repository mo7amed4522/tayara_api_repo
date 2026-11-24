package bxapi

import (
	"time"

	"github.com/mo7amed4522/tayara_api/model/bxapi/response"
	"github.com/mo7amed4522/tayara_api/utils"
)

const (
	sharedResponseTTL = time.Hour
	agentScopedTTL    = time.Minute
)

type cachedPage[T any] struct {
	Items     []T   `json:"items"`
	Total     int64 `json:"total"`
	Page      int   `json:"page"`
	PageSize  int   `json:"pageSize"`
	TotalPage int64 `json:"totalPage"`
}

func (page cachedPage[T]) paginateResult() *utils.PaginateResult {
	return &utils.PaginateResult{
		Total:     page.Total,
		Page:      page.Page,
		PageSize:  page.PageSize,
		TotalPage: page.TotalPage,
		Data:      page.Items,
	}
}

func applyWishlistFlags(pool []response.CustomerRecommendedResult, packageIDs []uint) {
	wishlisted := make(map[uint]struct{}, len(packageIDs))
	for _, id := range packageIDs {
		wishlisted[id] = struct{}{}
	}
	for i := range pool {
		_, pool[i].Wishlisted = wishlisted[pool[i].Package.ID]
	}
}
