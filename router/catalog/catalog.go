package catalog

import "github.com/gin-gonic/gin"

type CatalogRouter struct{}

// InitCatalogRouter registers package endpoints across the three guard groups:
// public (buyer browsing), manage (agent owns), system (admin moderates).
func (r CatalogRouter) InitCatalogRouter(publicGroup, manageGroup, systemGroup *gin.RouterGroup) {
	public := publicGroup.Group("api/v1/catalog/public")
	manage := manageGroup.Group("api/v1/catalog")
	system := systemGroup.Group("api/v1/catalog/admin")

	{
		public.GET("packages", catalogApi.ListPublicPackages)
		public.GET("packages/detail", catalogApi.GetPublicPackage)
		public.GET("packages/detail/reviews", catalogApi.ListPublicPackageReviews)
	}
	{
		manage.POST("packages", catalogApi.CreatePackage)
		manage.PUT("packages", catalogApi.UpdatePackage)
		manage.DELETE("packages", catalogApi.DeletePackage)
		manage.POST("packages/submit", catalogApi.SubmitForApproval)
		manage.GET("packages", catalogApi.ListOwnPackages)
		manage.GET("packages/detail", catalogApi.GetOwnPackage)
	}
	{
		system.GET("packages", catalogApi.AdminListPackages)
		system.GET("packages/detail", catalogApi.AdminGetPackageDetail)
		system.PUT("packages/status", catalogApi.AdminSetStatus)
	}
}
