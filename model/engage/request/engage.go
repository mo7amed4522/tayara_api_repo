package request

import (
	"time"

	commonReq "github.com/mo7amed4522/tayara_api/model/common/request"
)

// ReviewReq create/update a review & rating for a booking.
type ReviewReq struct {
	BookingID  uint   `json:"bookingId" binding:"required"`
	Rating     int    `json:"rating" binding:"required,min=1,max=5"`
	ReviewText string `json:"reviewText"`
}

// ReviewPageReq paginated review listing with optional agent/package filter.
type ReviewPageReq struct {
	commonReq.PageInfo
	AgentID   uint   `json:"agentId" form:"agentId"`
	PackageID uint   `json:"packageId" form:"packageId"`
	Status    string `json:"status" form:"status"`
}

// AdCreativeReq create/update an ad creative.
type AdCreativeReq struct {
	ID              uint       `json:"id"` // 0 on create
	Title           string     `json:"title"`
	Placement       string     `json:"placement" binding:"required"`
	DisplayOrder    int        `json:"displayOrder"`
	StartDatetime   *time.Time `json:"startDatetime"`
	EndDatetime     *time.Time `json:"endDatetime"`
	BannerImageURL  string     `json:"bannerImageURL" binding:"required"`
	OverlayTitle    string     `json:"overlayTitle"`
	OverlaySubtitle string     `json:"overlaySubtitle"`
	CtaLabel        string     `json:"ctaLabel"`
	RedirectType    string     `json:"redirectType"`
	TargetCountryID *uint      `json:"targetCountryId"`
	TargetAgentID   *uint      `json:"targetAgentId"`
	TargetPackageID *uint      `json:"targetPackageId"`
}

// AdCreativePageReq paginated ad listing with placement/status filter.
type AdCreativePageReq struct {
	commonReq.PageInfo
	Placement string `json:"placement" form:"placement"`
	Status    string `json:"status" form:"status"`
}

// AdSetStatusReq admin changes an ad's status (activate/disable).
type AdSetStatusReq struct {
	ID     uint   `json:"id" binding:"required"`
	Status string `json:"status" binding:"required"` // active | disabled
}

// AdClickReq records a click on a creative.
type AdClickReq struct {
	AdID       uint   `json:"adId" binding:"required"`
	SourcePage string `json:"sourcePage"`
}

// WishlistToggleReq add/remove a package from the user's wishlist.
type WishlistToggleReq struct {
	PackageID uint `json:"packageId" binding:"required"`
}

// ShareReq records a share of a package.
type ShareReq struct {
	PackageID uint   `json:"packageId" binding:"required"`
	Channel   string `json:"channel"`
}

// CouponReq create a coupon (admin).
type CouponReq struct {
	Code      string     `json:"code" binding:"required"`
	UserID    *uint      `json:"userId"`
	Value     int64      `json:"value"`
	ValidFrom *time.Time `json:"validFrom"`
	ValidTo   *time.Time `json:"validTo"`
}

// CouponPageReq paginated coupon listing.
type CouponPageReq struct {
	commonReq.PageInfo
	UserID uint   `json:"userId" form:"userId"`
	Status string `json:"status" form:"status"`
}

// NotificationTemplateReq upsert a notification template.
type NotificationTemplateReq struct {
	ID        uint   `json:"id"` // 0 on create
	EventType string `json:"eventType" binding:"required"`
	Channel   string `json:"channel" binding:"required"`
	Subject   string `json:"subject" binding:"required"`
	Body      string `json:"body"`
	IsActive  bool   `json:"isActive"`
}

// ResendNotificationReq re-emit an existing notification by id.
type ResendNotificationReq struct {
	ID uint `json:"id" binding:"required"`
}

// NotificationPageReq paginated notification listing with optional status filter.
type NotificationPageReq struct {
	commonReq.PageInfo
	Status string `json:"status" form:"status"`
}
