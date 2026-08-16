package request

import commonReq "github.com/mo7amed4522/tayara_api/model/common/request"

// CustomerSearchReq search-payload the mobile home search bar accepts.
type CustomerSearchReq struct {
	commonReq.PageInfo
	Keyword string `json:"keyword" form:"keyword" binding:"required"`
}

// CustomerHomeDealsReq paginated deals listing with optional category filter.
type CustomerHomeDealsReq struct {
	commonReq.PageInfo
	Category string `json:"category" form:"category"`
}

// CustomerHomeExploreReq paginated explore-trips listing with optional category filter.
type CustomerHomeExploreReq struct {
	commonReq.PageInfo
	Category string `json:"category" form:"category"`
}
