package engage

import "time"

// PushToken stores the active push token for one app installation and user.
type PushToken struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	UserID     uint      `json:"user_id" gorm:"not null;uniqueIndex:idx_user_device,priority:1"`
	DeviceID   string    `json:"device_id" gorm:"size:255;not null;uniqueIndex:idx_user_device,priority:2"`
	Platform   string    `json:"platform" gorm:"size:16;not null"`
	Token      string    `json:"token" gorm:"type:text;not null"`
	AppVersion string    `json:"app_version" gorm:"size:100"`
	LastSeenAt time.Time `json:"last_seen_at" gorm:"not null"`
	IsActive   bool      `json:"is_active" gorm:"not null;default:true"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (PushToken) TableName() string { return "push_tokens" }
