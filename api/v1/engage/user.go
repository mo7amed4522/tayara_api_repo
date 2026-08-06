package engage

import (
	"github.com/gin-gonic/gin"
	"github.com/mo7amed4522/tayara_api/model/common/response"
	engageReq "github.com/mo7amed4522/tayara_api/model/engage/request"
	"github.com/mo7amed4522/tayara_api/utils"
)

type UserEngageApi struct{}

// ToggleWishlist
// @Tags Engage
// @Summary User adds/removes a package from their wishlist
// @Produce application/json
// @Param data body engageReq.WishlistToggleReq true "wishlist payload"
// @Success 200 {object} response.Response
// @Router /api/v1/engage/wishlist [post]
func (UserEngageApi) ToggleWishlist(c *gin.Context) {
	userID := utils.GetBxUserID(c)
	var req engageReq.WishlistToggleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	on, err := engageService.UserEngageServices.ToggleWishlist(userID, req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(gin.H{"on": on}, c)
}

// ListWishlist
// @Tags Engage
// @Summary User lists their saved packages
// @Produce application/json
// @Success 200 {object} response.Response
// @Router /api/v1/engage/wishlist [get]
func (UserEngageApi) ListWishlist(c *gin.Context) {
	userID := utils.GetBxUserID(c)
	list, err := engageService.UserEngageServices.ListWishlist(userID)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(list, c)
}

// RecordShare
// @Tags Engage
// @Summary User records a share of a package
// @Produce application/json
// @Param data body engageReq.ShareReq true "share payload"
// @Success 200 {object} response.Response
// @Router /api/v1/engage/share [post]
func (UserEngageApi) RecordShare(c *gin.Context) {
	userID := utils.GetBxUserID(c)
	var req engageReq.ShareReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	share, err := engageService.UserEngageServices.RecordShare(userID, req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(share, c)
}

// CreateCoupon
// @Tags Engage
// @Summary Admin creates a coupon
// @Produce application/json
// @Param data body engageReq.CouponReq true "coupon payload"
// @Success 200 {object} response.Response
// @Router /api/v1/engage/admin/coupons [post]
func (UserEngageApi) CreateCoupon(c *gin.Context) {
	var req engageReq.CouponReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	coupon, err := engageService.UserEngageServices.CreateCoupon(req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(coupon, c)
}

// ListCoupons
// @Tags Engage
// @Summary User lists their coupons
// @Produce application/json
// @Param data query engageReq.CouponPageReq true "filter"
// @Success 200 {object} response.Response
// @Router /api/v1/engage/coupons [get]
func (UserEngageApi) ListCoupons(c *gin.Context) {
	userID := utils.GetBxUserID(c)
	var req engageReq.CouponPageReq
	_ = c.ShouldBindQuery(&req)
	list, total, err := engageService.UserEngageServices.ListCoupons(userID, req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(response.PageResult{List: list, Total: total, Page: req.Page, PageSize: req.PageSize}, "ok", c)
}

// AdminListCoupons
// @Tags Engage
// @Summary Admin lists all coupons
// @Produce application/json
// @Param data query engageReq.CouponPageReq true "filter"
// @Success 200 {object} response.Response
// @Router /api/v1/engage/admin/coupons [get]
func (UserEngageApi) AdminListCoupons(c *gin.Context) {
	var req engageReq.CouponPageReq
	_ = c.ShouldBindQuery(&req)
	list, total, err := engageService.UserEngageServices.ListCoupons(0, req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(response.PageResult{List: list, Total: total, Page: req.Page, PageSize: req.PageSize}, "ok", c)
}

// ApplyCoupon
// @Tags Engage
// @Summary User applies a coupon to a booking (discount + marks used)
// @Produce application/json
// @Param data body engageReq.ApplyCouponReq true "coupon apply payload"
// @Success 200 {object} response.Response
// @Router /api/v1/engage/coupons/apply [post]
func (UserEngageApi) ApplyCoupon(c *gin.Context) {
	userID := utils.GetBxUserID(c)
	var req engageReq.ApplyCouponReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	coupon, err := engageService.UserEngageServices.ApplyCoupon(userID, req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(coupon, c)
}

// RegisterDeviceToken
// @Tags Engage
// @Summary User registers a push device token (FCM)
// @Produce application/json
// @Param data body engageReq.DeviceTokenReq true "device token payload"
// @Success 200 {object} response.Response
// @Router /api/v1/engage/devices [post]
func (UserEngageApi) RegisterDeviceToken(c *gin.Context) {
	userID := utils.GetBxUserID(c)
	var req engageReq.DeviceTokenReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	token, err := engageService.UserEngageServices.RegisterDeviceToken(userID, req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(token, c)
}

// ListDeviceTokens
// @Tags Engage
// @Summary User lists their registered push devices
// @Produce application/json
// @Success 200 {object} response.Response
// @Router /api/v1/engage/devices [get]
func (UserEngageApi) ListDeviceTokens(c *gin.Context) {
	userID := utils.GetBxUserID(c)
	list, err := engageService.UserEngageServices.ListDeviceTokens(userID)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(list, c)
}

// RemoveDeviceToken
// @Tags Engage
// @Summary User deactivates a push device token
// @Produce application/json
// @Param data body engageReq.DeviceTokenReq true "device token payload"
// @Success 200 {object} response.Response
// @Router /api/v1/engage/devices/remove [post]
func (UserEngageApi) RemoveDeviceToken(c *gin.Context) {
	userID := utils.GetBxUserID(c)
	var req engageReq.DeviceTokenReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := engageService.UserEngageServices.RemoveDeviceToken(userID, req.Token); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("device removed", c)
}
