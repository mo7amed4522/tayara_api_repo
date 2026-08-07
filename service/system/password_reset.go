package system

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/model/system"
	systemReq "github.com/mo7amed4522/tayara_api/model/system/request"
	"github.com/mo7amed4522/tayara_api/service/engage/dispatch"
	"github.com/mo7amed4522/tayara_api/utils"
	"gorm.io/gorm"
)

// resetTokenTTL is the validity window (30 min) for a one-time password-reset token.
const resetTokenTTL = 30 * time.Minute

// ResetTokenLen is the generated token length in bytes before hex encoding.
const ResetTokenLen = 24

// RequestForgotPassword mints a one-time reset token for a system (admin) account and emails the
// reset link. The token is stored hashed. To avoid leaking whether an email is registered, this
// returns success even when the account does not exist.
func (sysAdminProfileServices *SysAdminProfileServices) RequestForgotPassword(req systemReq.SysForgotPasswordReq) error {
	if global.GVA_DB == nil {
		return errors.New("db not init")
	}
	var admin system.SysAdminProfile
	if global.GVA_DB.Where("email = ?", req.Email).First(&admin).Error != nil {
		return nil // identical response whether or not the account exists
	}
	raw, tokenHash, err := newResetToken()
	if err != nil {
		return err
	}
	expires := time.Now().Add(resetTokenTTL)
	if err := global.GVA_DB.Create(&system.PasswordReset{
		Email: req.Email, Role: "admin", TokenHash: tokenHash, ExpiresAt: &expires,
	}).Error; err != nil {
		return err
	}
	// SMTP delivery is best-effort; a down mailer never fails the request.
	var smtp dispatch.SmtpSender
	if smtp.Enabled() {
		_ = smtp.Send([]string{req.Email}, "Tayara password reset",
			"Use this one-time code to reset your password (valid 30 minutes):\n"+raw)
	}
	return nil
}

// ResetPassword validates a one-time reset token and sets the new password on the subject account.
func (s *SysAdminProfileServices) ResetPassword(req systemReq.SysResetPasswordReq) error {
	if global.GVA_DB == nil {
		return errors.New("db not init")
	}
	reset, err := findValidReset(req.Token, "admin")
	if err != nil {
		return err
	}
	return global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&system.SysAdminProfile{}).Where("email = ?", reset.Email).
			Update("password", utils.BcryptHash(req.NewPassword)).Error; err != nil {
			return err
		}
		return tx.Model(&system.PasswordReset{}).Where("id = ?", reset.ID).Update("used", true).Error
	})
}

// findValidReset returns an unused, non-expired reset record whose hash matches the raw token.
func findValidReset(rawToken, role string) (*system.PasswordReset, error) {
	var reset system.PasswordReset
	if err := global.GVA_DB.Where("token_hash = ? AND role = ? AND used = ?", hashToken(rawToken), role, false).
		Order("created_at desc").First(&reset).Error; err != nil {
		return nil, errors.New("invalid or expired reset token")
	}
	if reset.ExpiresAt == nil || time.Now().After(*reset.ExpiresAt) {
		return nil, errors.New("reset token expired")
	}
	return &reset, nil
}

func newResetToken() (raw, hash string, err error) {
	buf := make([]byte, ResetTokenLen)
	if _, err = rand.Read(buf); err != nil {
		return "", "", err
	}
	raw = hex.EncodeToString(buf)
	return raw, hashToken(raw), nil
}

func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
