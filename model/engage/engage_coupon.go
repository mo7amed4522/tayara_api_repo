package engage

import (
	"time"

	"github.com/mo7amed4522/tayara_api/global"
)

// Coupon a promo/gift code a user holds. Display-scope for MVP: records code + value + validity;
// applying it to a booking is a later phase.
type Coupon struct {
	ID        uint       `json:"id" gorm:"primarykey;comment:ID"`
	Code      string     `json:"code" form:"code" gorm:"size:64;uniqueIndex;comment:Coupon code"`
	UserID    *uint      `json:"userId" form:"userId" gorm:"index;comment:Owning user ID"`
	Value     int64      `json:"value" form:"value" gorm:"default:0;comment:Value, minor units"`
	Status    string     `json:"status" form:"status" gorm:"size:32;index;default:active;comment:active|used|expired"`
	ValidFrom *time.Time `json:"validFrom" form:"validFrom" gorm:"comment:Valid from"`
	ValidTo   *time.Time `json:"validTo" form:"validTo" gorm:"comment:Valid to"`
	global.GVA_MODEL
}

func (Coupon) TableName() string { return "engage_coupon" }
