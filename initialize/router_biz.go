package initialize

import (
	"github.com/gin-gonic/gin"
	"github.com/mo7amed4522/tayara_api/middleware"
	"github.com/mo7amed4522/tayara_api/router"
)

func holder(routers ...*gin.RouterGroup) {
	_ = routers
}
func initBizRouter(routers ...*gin.RouterGroup) {
	privateGroup := routers[0]
	publicGroup := routers[1]
	bxPrivateGroup := routers[2]
	holder(publicGroup, privateGroup, bxPrivateGroup)

	bxPrivate := publicGroup.Group("")
	bxPrivate.Use(middleware.BxDomainJWTAuth())
	managePrivate := publicGroup.Group("")
	managePrivate.Use(middleware.ManageJWTAuth())
	systemPrivate := publicGroup.Group("")
	systemPrivate.Use(middleware.SystemJWTAuth())

	// admin
	router.RouterGroupApp.System.InitSysAdminProfileRouter(publicGroup, systemPrivate)

	// catalog (packages by agent, moderation by admin, browsing by public/buyer)
	router.RouterGroupApp.Catalog.InitCatalogRouter(publicGroup, managePrivate, systemPrivate)
	router.RouterGroupApp.Catalog.InitCustomTripRouter(managePrivate)

	// booking lifecycle (user requests, agent reviews/offers, admin monitors/force-cancels)
	router.RouterGroupApp.Booking.InitBookingRouter(publicGroup, managePrivate, systemPrivate)

	// payment (user opens payment + Stripe webhook; admin captures/refunds; agent payout views)
	router.RouterGroupApp.Payment.InitPaymentRouter(publicGroup, managePrivate, systemPrivate)

	// engage (reviews, ads, notifications, wishlist/shares/coupons)
	router.RouterGroupApp.Engage.InitEngageRouter(publicGroup, systemPrivate)

	// analytics (admin aggregate cards + agent dashboard card)
	router.RouterGroupApp.Analytics.InitAnalyticsRouter(managePrivate, systemPrivate)

}
