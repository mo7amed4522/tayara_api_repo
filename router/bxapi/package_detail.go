package bxapi

import "github.com/gin-gonic/gin"

type PackageDetailRouter struct{}

// InitPackageDetailRouter registers customer-facing package detail endpoints.
func (r PackageDetailRouter) InitPackageDetailRouter(publicGroup, bxGroup *gin.RouterGroup) {
	customer := publicGroup.Group("api/v1/customer")
	bxCustomer := bxGroup.Group("api/v1/customer")

	{
		customer.GET("packages/detail", packageDetailApi.GetPackageDetail)
		customer.GET("packages/detail/reviews", packageDetailApi.ListPackageReviews)
	}
	{
		bxCustomer.GET("packages/detail", packageDetailApi.GetPackageDetail)
		bxCustomer.GET("packages/detail/reviews", packageDetailApi.ListPackageReviews)
	}
}
