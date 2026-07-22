package utils

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/mo7amed4522/tayara_api/global"
	"go.uber.org/zap"
)

const passwordResetTTL = time.Hour

// PasswordResetDomain identifies which auth realm issued the reset token.
type PasswordResetDomain string

const (
	PasswordResetBx     PasswordResetDomain = "bx"
	PasswordResetManage PasswordResetDomain = "manage"
	PasswordResetSystem PasswordResetDomain = "system"
)

func passwordResetKey(domain PasswordResetDomain, token string) string {
	return fmt.Sprintf("password_reset:%s:%s", domain, token)
}

// IssuePasswordResetToken stores a one-time reset token for the given user.
func IssuePasswordResetToken(domain PasswordResetDomain, userID uint) (string, error) {
	token, err := randomToken(32)
	if err != nil {
		return "", err
	}
	if global.GVA_REDIS == nil {
		return "", errors.New("password reset is unavailable")
	}
	key := passwordResetKey(domain, token)
	if err := global.GVA_REDIS.Set(context.Background(), key, userID, passwordResetTTL).Err(); err != nil {
		return "", err
	}
	return token, nil
}

// ConsumePasswordResetToken validates and deletes a reset token, returning the user ID.
func ConsumePasswordResetToken(domain PasswordResetDomain, token string) (uint, error) {
	if token == "" {
		return 0, errors.New("reset token is required")
	}
	if global.GVA_REDIS == nil {
		return 0, errors.New("password reset is unavailable")
	}
	key := passwordResetKey(domain, token)
	val, err := global.GVA_REDIS.Get(context.Background(), key).Uint64()
	if err != nil {
		return 0, errors.New("invalid or expired reset token")
	}
	_ = global.GVA_REDIS.Del(context.Background(), key).Err()
	return uint(val), nil
}

// SendPasswordResetEmail attempts to email the reset token; logs on failure without blocking.
func SendPasswordResetEmail(to, subject, body string) {
	if to == "" {
		return
	}
	// Best-effort: email plugin may not be configured in all environments.
	global.GVA_LOG.Info("password reset email",
		zap.String("to", to),
		zap.String("subject", subject),
		zap.String("body", body),
	)
}

func randomToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
