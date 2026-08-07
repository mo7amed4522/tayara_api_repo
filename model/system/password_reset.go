package system

import (
	"time"

	"github.com/mo7amed4522/tayara_api/global"
)

// PasswordReset a one-time password-reset token for an admin or agent. The token is stored
// hashed; it is single-use and expires after ResetTokenTTL. Email (or phone for agents) is the
// subject of the reset so a leaked token isn't enough — recovery also needs mailbox/phone access.
type PasswordReset struct {
	ID        uint       `json:"id" gorm:"primarykey;comment:ID"`
	Email     string     `json:"email" form:"email" gorm:"size:191;index;comment:Subject email"`
	Role      string     `json:"role" form:"role" gorm:"size:16;index;comment:admin|agent"`
	TokenHash string     `json:"-" gorm:"size:128;comment:hashed one-time token"`
	Used      bool       `json:"used" form:"used" gorm:"default:false;comment:Consumed"`
	ExpiresAt *time.Time `json:"expiresAt" form:"expiresAt" gorm:"comment:Token expiry"`
	global.GVA_MODEL
}

func (PasswordReset) TableName() string { return "password_reset" }
