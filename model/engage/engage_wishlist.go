package engage

import "github.com/mo7amed4522/tayara_api/global"

// Wishlist a user's saved packages (toggle on/off).
type Wishlist struct {
	ID        uint `json:"id" gorm:"primarykey;comment:ID"`
	UserID    uint `json:"userId" form:"userId" gorm:"uniqueIndex:uniq_wish_user_pkg;comment:User ID"`
	PackageID uint `json:"packageId" form:"packageId" gorm:"uniqueIndex:uniq_wish_user_pkg;comment:Package ID"`
	global.GVA_MODEL
}

func (Wishlist) TableName() string { return "engage_wishlist" }
