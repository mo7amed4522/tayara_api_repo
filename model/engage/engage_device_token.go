package engage

import "github.com/mo7amed4522/tayara_api/global"

// DeviceToken a user's registered push token (FCM) for notification delivery. A user may have
// several devices; each token is unique. Admin can list an end-user's devices to manage pushes.
type DeviceToken struct {
	ID       uint   `json:"id" gorm:"primarykey;comment:ID"`
	UserID   uint   `json:"userId" form:"userId" gorm:"index;comment:Owning user ID"`
	Token    string `json:"token" form:"token" gorm:"size:500;uniqueIndex;comment:FCM device token"`
	Platform string `json:"platform" form:"platform" gorm:"size:32;index;comment:ios|android|web"`
	IsActive bool   `json:"isActive" form:"isActive" gorm:"default:true;comment:Is active"`
	global.GVA_MODEL
}

func (DeviceToken) TableName() string { return "engage_device_token" }
