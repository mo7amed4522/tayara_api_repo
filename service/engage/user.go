package engage

import (
	"errors"
	"time"

	"github.com/mo7amed4522/tayara_api/constants"
	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/model/engage"
	engageReq "github.com/mo7amed4522/tayara_api/model/engage/request"
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
func (u UserEngageServices) ListCoupons(userID uint, req engageReq.CouponPageReq) ([]engage.Coupon, int64, error) {
	if global.GVA_DB == nil {
		return nil, 0, errors.New("db not init")
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
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []engage.Coupon
	if err := db.Scopes(req.Paginate()).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
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
