package push

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
	"time"

	"github.com/mo7amed4522/tayara_api/constants"
	"github.com/mo7amed4522/tayara_api/global"
	supportModel "github.com/mo7amed4522/tayara_api/model/support"
	"github.com/mo7amed4522/tayara_api/service/base"
)

func uintString(v uint) string { return strconv.FormatUint(uint64(v), 10) }

// sha256Hex is the idempotency/token hash primitive.
func sha256Hex(parts ...string) string {
	h := sha256.New()
	for _, p := range parts {
		_, _ = h.Write([]byte(p))
	}
	return hex.EncodeToString(h.Sum(nil))
}

// IdempotencyKey is sha256(message_uuid + recipient_actor_type + recipient_actor_id + token).
func IdempotencyKey(messageUUID, recipientType string, recipientID uint, token string) string {
	return sha256Hex(messageUUID, recipientType, uintString(recipientID), token)
}

// maxAttempts caps in-process and sweeper retries; it reuses the support.kafka.retry shape.
func maxAttempts() int {
	n := global.GVA_CONFIG.Support.Push.Retry.MaxAttempts
	if n <= 0 {
		n = 3
	}
	return n
}

// backoffFor reuses support.kafka.retry's backoff values for push retries.
func backoffFor(attempt int) time.Duration {
	cfg := global.GVA_CONFIG.Support.Kafka.Retry
	initial := cfg.InitialBackoffMS
	if initial <= 0 {
		initial = 200
	}
	max := cfg.MaxBackoffMS
	if max <= 0 {
		max = 5000
	}
	d := time.Duration(initial) * time.Millisecond
	for i := 1; i < attempt; i++ {
		d *= 2
		if d >= time.Duration(max)*time.Millisecond {
			break
		}
	}
	if ceiling := time.Duration(max) * time.Millisecond; d > ceiling {
		d = ceiling
	}
	return d
}

func collapseWindow() time.Duration {
	secs := global.GVA_CONFIG.Support.Push.CollapseWindowSeconds
	if secs <= 0 {
		secs = 30
	}
	return time.Duration(secs) * time.Second
}

func previewBytes() int {
	n := global.GVA_CONFIG.Support.Push.MaxPreviewBytes
	if n <= 0 {
		n = 140
	}
	return n
}

func titleBytes() int {
	n := global.GVA_CONFIG.Support.Push.MaxTitleBytes
	if n <= 0 {
		n = 60
	}
	return n
}

// silentRole reports whether this role should get a silent push when presence is unknown.
func silentRole(actorType string) bool {
	roles := global.GVA_CONFIG.Support.Push.SilentForRoles
	if len(roles) == 0 {
		roles = []string{constants.ActorSupport, constants.ActorQuality}
	}
	for _, r := range roles {
		if strings.TrimSpace(r) == actorType {
			return true
		}
	}
	return false
}

// Enabled reads the runtime override first (base_system_setting, section `support`) and falls
// back to the YAML default. Anything an admin can toggle goes through that store.
func Enabled() bool {
	fallback := global.GVA_CONFIG.Support.Push.Enabled
	if global.GVA_DB == nil {
		return fallback
	}
	st, err := (base.LookupServices{}).GetSetting(constants.SupportSettingPushEnabled)
	if err != nil || st == nil {
		return fallback
	}
	b, err := strconv.ParseBool(strings.TrimSpace(st.Value))
	if err != nil {
		return fallback
	}
	return b
}

// NotificationCode is the deterministic engage_notification code for one (message, recipient)
// pair. The unique index on notification_code is what makes the notification log idempotent under
// Kafka at-least-once redelivery, without a second ledger.
func NotificationCode(messageUUID, recipientKey string) string {
	return constants.CodePrefixNotification + strings.ToUpper(sha256Hex(messageUUID, "|", recipientKey)[:24])
}

// revokeTokenRow appends a revocation row for a token FCM rejected. Never a DELETE/UPDATE.
func revokeTokenRow(token string) error {
	if global.GVA_DB == nil || token == "" {
		return nil
	}
	var latest supportModel.PushToken
	if err := global.GVA_DB.Where("token_hash = ?", sha256Hex(token)).Order("id desc").First(&latest).Error; err != nil {
		return err
	}
	if latest.RevokedAt != nil {
		return nil
	}
	now := time.Now().UTC()
	return global.GVA_DB.Create(&supportModel.PushToken{
		ActorType: latest.ActorType, ActorID: latest.ActorID, Platform: latest.Platform,
		Token: latest.Token, TokenHash: latest.TokenHash, AppVersion: latest.AppVersion,
		RevokedAt: &now,
	}).Error
}

