package engage

import "github.com/gin-gonic/gin"

type EngageRouter struct{}

// InitEngageRouter registers the content & engagement endpoints across the guard groups.
func (r EngageRouter) InitEngageRouter(publicGroup, bxGroup, systemGroup *gin.RouterGroup) {
	public := publicGroup.Group("api/v1/engage")
	bx := bxGroup.Group("api/v1/engage")
	system := systemGroup.Group("api/v1/engage")

	// public (browse)
	{
		public.GET("reviews", reviewApi.ListReviews)
		public.GET("ads", adApi.ActiveAds)
		public.POST("ads/impression", adApi.RecordImpression)
		public.POST("ads/click", adApi.RecordClick)
	}
	// authenticated buyer (bx token): reviews, wishlist, shares, coupons, notifications, devices
	{
		bx.POST("reviews", reviewApi.CreateReview)
		bx.POST("wishlist", userApi.ToggleWishlist)
		bx.GET("wishlist", userApi.ListWishlist)
		bx.POST("share", userApi.RecordShare)
		bx.GET("coupons", userApi.ListCoupons)
		bx.POST("coupons/apply", userApi.ApplyCoupon)
		bx.GET("notifications", notifApi.ListMyNotifications)
		bx.POST("devices", userApi.RegisterDeviceToken)
		bx.GET("devices", userApi.ListDeviceTokens)
		bx.POST("devices/remove", userApi.RemoveDeviceToken)
	}
	// admin-facing
	{
		system.GET("admin/reviews", reviewApi.AdminListReviews)
		system.PUT("admin/reviews/status", reviewApi.SetStatus)
		system.POST("ads", adApi.CreateAd)
		system.PUT("ads", adApi.UpdateAd)
		system.PUT("ads/status", adApi.SetStatus)
		system.DELETE("ads", adApi.DeleteAd)
		system.GET("admin/ads", adApi.ListAds)
		system.GET("admin/ads/analytics", adApi.GroupedAnalytics)
		system.GET("admin/notifications", notifApi.ListNotifications)
		system.POST("admin/notifications/resend", notifApi.Resend)
		system.POST("admin/notification-templates", notifApi.UpsertTemplate)
		system.GET("admin/notification-templates", notifApi.ListTemplates)
		system.DELETE("admin/notification-templates", notifApi.DeleteTemplate)
		system.POST("admin/coupons", userApi.CreateCoupon)
		system.GET("admin/coupons", userApi.AdminListCoupons)
	}
}
