package request

import commonReq "github.com/mo7amed4522/tayara_api/model/common/request"

// CustomerExploreTopicsReq represents the request to list explore topics
type CustomerExploreTopicsReq struct {
	commonReq.PageInfo
}

// CustomerExploreTopicPackagesReq represents the request for packages in a topic
type CustomerExploreTopicPackagesReq struct {
	commonReq.PageInfo
	Category string `json:"category" form:"category"`
}