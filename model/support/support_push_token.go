package support

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// PushToken is an append-only device registration. Revocation is a NEW row with revoked_at set,
// never an UPDATE, mirroring the rest of the support module's append-only tables. The active
// token set is "the latest row per (actor_type, actor_id, platform) whose revoked_at is NULL".
type PushToken struct {
	ID         uint       `json:"id" gorm:"primarykey;comment:ID"`
	ActorType  string     `json:"actorType" form:"actorType" gorm:"size:16;index:idx_push_actor;comment:customer|support|quality|admin"`
	ActorID    uint       `json:"actorId" form:"actorId" gorm:"index:idx_push_actor;comment:bx_user_profile.id or sys_admin_profile.id"`
	Platform   string     `json:"platform" form:"platform" gorm:"size:16;comment:ios|android|web"`
	Token      string     `json:"-" form:"token" gorm:"size:512;comment:Device push token"`
	TokenHash  string     `json:"tokenHash" form:"tokenHash" gorm:"size:64;index;comment:sha256(token) - dedupe key"`
	AppVersion string     `json:"appVersion" form:"appVersion" gorm:"size:32;comment:Client app version"`
	RevokedAt  *time.Time `json:"revokedAt" form:"revokedAt" gorm:"comment:Set on a revocation row"`
	CreatedAt  time.Time  `json:"createdAt" gorm:"autoCreateTime;<-:create;comment:Server created at"`
}

func (PushToken) TableName() string { return "support_push_token" }

func (PushToken) BeforeUpdate(tx *gorm.DB) error {
	return errors.New("support_push_token is append-only")
}
func (PushToken) BeforeDelete(tx *gorm.DB) error {
	return errors.New("support_push_token is append-only")
}
