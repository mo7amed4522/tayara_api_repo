package response

// CustomerExploreTopicCard represents a topic card for the explore page
type CustomerExploreTopicCard struct {
	ID           uint   `json:"id"`
	Code         string `json:"code"`
	Name         string `json:"name"`
	NameAr       string `json:"nameAr"`
	Description  string `json:"description"`
	ImageURL     string `json:"imageUrl"`
	IconURL      string `json:"iconUrl"`
	PackageCount int    `json:"packageCount"`
}

// CustomerExploreTopicDetail represents a topic with featured packages
type CustomerExploreTopicDetail struct {
	CustomerExploreTopicCard
	FeaturedPackages []CustomerExploreTripCard `json:"featuredPackages"`
}

// CustomerExploreTopicPackagesResult represents paginated packages for a topic
type CustomerExploreTopicPackagesResult struct {
	Topic    CustomerExploreTopicCard         `json:"topic"`
	Packages []CustomerExploreTripCard        `json:"packages"`
	Total    int64                            `json:"total"`
	Page     int                              `json:"page"`
	PageSize int                              `json:"pageSize"`
}

// AgentExploreTopicCard represents a topic card for agent dashboard (includes my package count)
type AgentExploreTopicCard struct {
	CustomerExploreTopicCard
	MyPackageCount int `json:"myPackageCount"`
}