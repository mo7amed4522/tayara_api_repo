package request

import commonReq "github.com/mo7amed4522/tayara_api/model/common/request"

// ExploreTopicReq represents the request to create/update an explore topic
type ExploreTopicReq struct {
	Code        string `json:"code" form:"code" binding:"required,alphanum,max=64"`
	Name        string `json:"name" form:"name" binding:"required,max=128"`
	NameAr      string `json:"nameAr" form:"nameAr" binding:"max=128"`
	Description string `json:"description" form:"description" binding:"max=500"`
	ImageURL    string `json:"imageUrl" form:"imageUrl" binding:"url,max=500"`
	IconURL     string `json:"iconUrl" form:"iconUrl" binding:"url,max=500"`
	SortOrder   int    `json:"sortOrder" form:"sortOrder"`
	IsActive    bool   `json:"isActive" form:"isActive"`
}

// AdminExploreTopicPageReq represents the paginated request for admin topic listing
type AdminExploreTopicPageReq struct {
	commonReq.PageInfo
	IsActive *bool  `json:"isActive" form:"isActive"`
	Keyword  string `json:"keyword" form:"keyword"`
}

// TopicOrderReq represents a topic sort order update request
type TopicOrderReq struct {
	ID        uint `json:"id" binding:"required"`
	SortOrder int  `json:"sortOrder" binding:"required"`
}

// AgentPackageTopicsReq represents the request to associate a package with topics
type AgentPackageTopicsReq struct {
	TopicIDs []uint `json:"topicIds" binding:"required,min=1"`
}

// AdminManageTopicPackagesReq represents the request to bulk manage packages in a topic
type AdminManageTopicPackagesReq struct {
	PackageIDs []uint `json:"packageIds" binding:"required,min=1"`
	Replace    bool   `json:"replace"` // true = replace all, false = add only
}