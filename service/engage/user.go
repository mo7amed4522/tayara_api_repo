package engage

import (
	"errors"
	"time"

	"github.com/mo7amed4522/tayara_api/constants"
	"github.com/mo7amed4522/tayara_api/global"
	bookingModel "github.com/mo7amed4522/tayara_api/model/booking"
	"github.com/mo7amed4522/tayara_api/model/engage"
	engageReq "github.com/mo7amed4522/tayara_api/model/engage/request"
	"github.com/mo7amed4522/tayara_api/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UserEngageServices struct{}

// --- wishlist ---

// ToggleWishlist adds or removes a package from the user's wishlist (returns on = added).
func (u UserEngageServices) ToggleWishlist(userID uint, req engageReq.WishlistToggleReq) (bool, error) {
	if global.GVA_DB == nil {
		return false, errors.New("db not init")
	}
	var w engage.Wishlist
	err := global.GVA_DB.Where("user_id = ? AND package_id = ?", userID, req.PackageID).First(&w).Error
	if err == nil {
		// present -> remove
		return false, global.GVA_DB.Delete(&engage.Wishlist{}, w.ID).Error
	}
	// absent -> add
	if err := global.GVA_DB.Clauses(clause.OnConflict{DoNothing: true}).Create(&engage.Wishlist{
		UserID: userID, PackageID: req.PackageID,
	}).Error; err != nil {
		return false, err
	}
	return true, nil
}

// ListWishlist returns the user's saved packages (ids).
func (u UserEngageServices) ListWishlist(userID uint) ([]engage.Wishlist, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	var list []engage.Wishlist
	if err := global.GVA_DB.Where("user_id = ?", userID).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// --- shares ---

// RecordShare logs a share event (idempotent per user+package+channel).
func (u UserEngageServices) RecordShare(userID uint, req engageReq.ShareReq) (*engage.Share, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	if req.Channel == "" {
		req.Channel = "default"
	}
	now := time.Now()
	share := &engage.Share{UserID: userID, PackageID: req.PackageID, Channel: req.Channel, SharedAt: &now}
	err := global.GVA_DB.Clauses(clause.OnConflict{DoNothing: true}).Create(share).Error
	if err != nil {
		return nil, err
	}
	return share, nil
}

// --- coupons (display-scope for MVP) ---

// CreateCoupon creates a coupon (admin).
func (u UserEngageServices) CreateCoupon(req engageReq.CouponReq) (*engage.Coupon, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	coupon := &engage.Coupon{
		Code: req.Code, UserID: req.UserID, Value: req.Value,
		Status: constants.CouponStatusActive, ValidFrom: req.ValidFrom, ValidTo: req.ValidTo,
	}
	if err := global.GVA_DB.Create(coupon).Error; err != nil {
		return nil, err
	}
	return coupon, nil
}

// ListCoupons lists coupons (admin: all; user: scoped by userID).
func (u UserEngageServices) ListCoupons(userID uint, req engageReq.CouponPageReq) ([]engage.Coupon, int64, int64, error) {
	if global.GVA_DB == nil {
		return nil, 0, 0, errors.New("db not init")
	}
	db := global.GVA_DB.Model(&engage.Coupon{})
	if userID > 0 {
		db = db.Where("user_id = ?", userID)
	}
	if req.UserID > 0 {
		db = db.Where("user_id = ?", req.UserID)
	}
	if req.Status != "" {
		db = db.Where("status = ?", req.Status)
	}
	var list []engage.Coupon
	pr, err := utils.PaginateWithCustomQuery(db, db.Order("created_at desc"), req.Page, req.PageSize, &list)
	if err != nil {
		return nil, 0, 0, err
	}
	return list, pr.Total, pr.TotalPage, nil
}

// GetUserCoupon returns a single coupon a user owns (by code).
func (u UserEngageServices) GetUserCoupon(userID uint, code string) (*engage.Coupon, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	var c engage.Coupon
	if err := global.GVA_DB.Where("code = ? AND user_id = ?", code, userID).First(&c).Error; err != nil {
		return nil, errors.New("coupon not found")
	}
	return &c, nil
}

// ApplyCoupon validates a user's coupon and applies it as a discount to a booking. Rules: the
// coupon must be active, owned by the user, within its validity window, and not exceed the booking
// amount. On success the coupon is marked used and the booking's discount is updated.
func (u UserEngageServices) ApplyCoupon(userID uint, req engageReq.ApplyCouponReq) (*engage.Coupon, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	var coupon engage.Coupon
	if err := global.GVA_DB.Where("code = ? AND user_id = ?", req.Code, userID).First(&coupon).Error; err != nil {
		return nil, errors.New("coupon not found or not owned by this user")
	}
	if coupon.Status != constants.CouponStatusActive {
		return nil, errors.New("coupon is not active")
	}
	now := time.Now()
	if coupon.ValidFrom != nil && now.Before(*coupon.ValidFrom) {
		return nil, errors.New("coupon is not yet valid")
	}
	if coupon.ValidTo != nil && now.After(*coupon.ValidTo) {
		_ = global.GVA_DB.Model(&engage.Coupon{}).Where("id = ?", coupon.ID).
			Update("status", constants.CouponStatusExpired).Error
		return nil, errors.New("coupon has expired")
	}
	var bk bookingModel.Booking
	if err := global.GVA_DB.Where("id = ? AND user_id = ?", req.BookingID, userID).First(&bk).Error; err != nil {
		return nil, errors.New("booking not found")
	}
	if bk.DiscountAmount > 0 {
		return nil, errors.New("a discount is already applied to this booking")
	}
	if coupon.Value > bk.AmountGross {
		return nil, errors.New("coupon value exceeds the booking amount")
	}
	// apply discount + mark used atomically
	err := global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&bookingModel.Booking{}).Where("id = ?", bk.ID).
			Updates(map[string]interface{}{"discount_amount": coupon.Value, "coupon_code": req.Code}).Error; err != nil {
			return err
		}
		return tx.Model(&engage.Coupon{}).Where("id = ?", coupon.ID).
			Update("status", constants.CouponStatusUsed).Error
	})
	if err != nil {
		return nil, err
	}
	coupon.Status = constants.CouponStatusUsed
	return &coupon, nil
}

// RegisterDeviceToken upserts a user's FCM device token (idempotent on token value).
func (u UserEngageServices) RegisterDeviceToken(userID uint, req engageReq.DeviceTokenReq) (*engage.DeviceToken, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	token := &engage.DeviceToken{UserID: userID, Token: req.Token, Platform: req.Platform, IsActive: true}
	if err := global.GVA_DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "token"}},
		DoUpdates: clause.AssignmentColumns([]string{"user_id", "platform", "is_active"}),
	}).Create(token).Error; err != nil {
		return nil, err
	}
	return token, nil
}

// ListDeviceTokens returns a user's registered push tokens (admin can view an end-user's devices).
func (u UserEngageServices) ListDeviceTokens(userID uint) ([]engage.DeviceToken, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	var list []engage.DeviceToken
	if err := global.GVA_DB.Where("user_id = ?", userID).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// RemoveDeviceToken deactivates a user's device token (e.g. on logout).
func (u UserEngageServices) RemoveDeviceToken(userID uint, token string) error {
	if global.GVA_DB == nil {
		return errors.New("db not init")
	}
	return global.GVA_DB.Model(&engage.DeviceToken{}).
		Where("user_id = ? AND token = ?", userID, token).
		Update("is_active", false).Error
}
