package response

import "github.com/mo7amed4522/tayara_api/model/bxapi/response"

// ExploreTopicCard represents a topic card for listing
type ExploreTopicCard struct {
	ID           uint   `json:"id"`
	Code         string `json:"code"`
	Name         string `json:"name"`
	NameAr       string `json:"nameAr"`
	Description  string `json:"description"`
	ImageURL     string `json:"imageUrl"`
	IconURL      string `json:"iconUrl"`
	PackageCount int    `json:"packageCount"`
}

// ExploreTopicDetail represents a topic with featured packages
type ExploreTopicDetail struct {
	ExploreTopicCard
	FeaturedPackages []response.CustomerExploreTripCard `json:"featuredPackages"`
}

// ExploreTopicPackagesResult represents paginated packages for a topic
type ExploreTopicPackagesResult struct {
	Topic    ExploreTopicCard                    `json:"topic"`
	Packages []response.CustomerExploreTripCard  `json:"packages"`
	Total    int64                               `json:"total"`
	Page     int                                 `json:"page"`
	PageSize int                                 `json:"pageSize"`
}

// AgentTopicCard represents a topic card for agent dashboard (includes my package count)
type AgentTopicCard struct {
	ExploreTopicCard
	MyPackageCount int `json:"myPackageCount"`
}

// AdminExploreTopicDetail represents a topic for admin management
type AdminExploreTopicDetail struct {
	ID          uint   `json:"id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	NameAr      string `json:"nameAr"`
	Description string `json:"description"`
	ImageURL    string `json:"imageUrl"`
	IconURL     string `json:"iconUrl"`
	SortOrder   int    `json:"sortOrder"`
	IsActive    bool   `json:"isActive"`
	PackageCount int   `json:"packageCount"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

// AdminExploreTopicListResult represents the paginated admin topic list result
type AdminExploreTopicListResult struct {
	List       []AdminExploreTopicDetail `json:"list"`
	Total      int64                     `json:"total"`
	Page       int                       `json:"page"`
	PageSize   int                       `json:"pageSize"`
	TotalPages int64                     `json:"totalPages"`
}