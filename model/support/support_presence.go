package support

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// Presence an append-only "last seen" row. Never updated or deleted; the latest row per actor
// (MAX(id)) is the reported last_seen_at. Deliberately not a column on the profile tables.
type Presence struct {
	ID         uint      `json:"id" gorm:"primarykey;comment:ID"`
	ActorType  string    `json:"actorType" form:"actorType" gorm:"size:16;index:idx_actor;comment:customer|support|quality|admin"`
	ActorID    uint      `json:"actorId" form:"actorId" gorm:"index:idx_actor;comment:bx_user_profile.id or sys_admin_profile.id"`
	LastSeenAt time.Time `json:"lastSeenAt" form:"lastSeenAt" gorm:"index;comment:Last seen at (server clock, UTC)"`
	CreatedAt  time.Time `json:"createdAt" gorm:"autoCreateTime;<-:create;comment:Server created at"`
}

func (Presence) TableName() string { return "support_presence" }

func (Presence) BeforeUpdate(tx *gorm.DB) error { return errors.New("support_presence is append-only") }
func (Presence) BeforeDelete(tx *gorm.DB) error { return errors.New("support_presence is append-only") }
