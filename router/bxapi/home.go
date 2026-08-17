package bxapi

import "github.com/gin-gonic/gin"

type HomeRouter struct{}

// InitHomeRouter registers public buyer-facing home endpoints (and one authenticated endpoint).
// All routes live under /api/v1/customer on the public group.
func (r HomeRouter) InitHomeRouter(publicGroup, bxGroup *gin.RouterGroup) {
	customer := publicGroup.Group("api/v1/customer")
	bxCustomer := bxGroup.Group("api/v1/customer")

	// Public: search, explore, destinations, deals, agents, ads
	{
		customer.GET("search", homeApi.CustomerSearch)
		customer.GET("explore/trips", homeApi.ListExploreTrips)
		customer.GET("home/destinations", homeApi.ListDestinations)
		customer.GET("home/deals", homeApi.ListDeals)
		customer.GET("home/agents", homeApi.ListTopRatedAgents)
		customer.GET("home/ads", homeApi.ActiveHomeAds)
	}
	// Authenticated: recommended for you (bx token)
	{
		bxCustomer.GET("home/recommended", homeApi.RecommendedForYou)
	}
	// Package detail (duplicated on bx group for authenticated path)
	{
		customer.GET("packages/detail", packageDetailApi.GetPackageDetail)
		customer.GET("packages/detail/reviews", packageDetailApi.ListPackageReviews)
		bxCustomer.GET("packages/detail", packageDetailApi.GetPackageDetail)
		bxCustomer.GET("packages/detail/reviews", packageDetailApi.ListPackageReviews)
	}
}
