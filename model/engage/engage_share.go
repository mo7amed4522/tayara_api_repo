package engage

import (
	"time"

	"github.com/mo7amed4522/tayara_api/global"
)

// Share a share event of a package (channel: whatsapp/instagram/etc.).
type Share struct {
	ID        uint       `json:"id" gorm:"primarykey;comment:ID"`
	UserID    uint       `json:"userId" form:"userId" gorm:"uniqueIndex:uniq_share_user_pkg_channel;comment:User ID"`
	PackageID uint       `json:"packageId" form:"packageId" gorm:"uniqueIndex:uniq_share_user_pkg_channel;comment:Package ID"`
	Channel   string     `json:"channel" form:"channel" gorm:"size:32;uniqueIndex:uniq_share_user_pkg_channel;comment:Channel"`
	SharedAt  *time.Time `json:"sharedAt" form:"sharedAt" gorm:"comment:Shared at"`
	global.GVA_MODEL
}

func (Share) TableName() string { return "engage_share" }
