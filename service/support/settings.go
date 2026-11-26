package support

import (
	"errors"
	"strconv"
	"strings"

	"github.com/mo7amed4522/tayara_api/constants"
	"github.com/mo7amed4522/tayara_api/global"
	basemodel "github.com/mo7amed4522/tayara_api/model/base"
	baseReq "github.com/mo7amed4522/tayara_api/model/base/request"
	supportReq "github.com/mo7amed4522/tayara_api/model/support/request"
	"github.com/mo7amed4522/tayara_api/service/base"
)

// SettingsServices reads and writes runtime chat policy. The DB value (base_system_setting,
// section `support`) wins over the YAML default, exactly like the `travel.*` settings.
type SettingsServices struct{}

// storageMaxFileBytes mirrors utils/upload/local.go's hard 10 MiB cap. A configured value can
// only tighten it, never exceed it.
const storageMaxFileBytes int64 = 10 << 20

// value returns the runtime setting, falling back to the YAML-derived default.
func (s SettingsServices) value(key, fallback string) string {
	if global.GVA_DB != nil {
		var st basemodel.SystemSetting
		if err := global.GVA_DB.Where("key = ? AND is_active = ?", key, true).First(&st).Error; err == nil {
			if v := strings.TrimSpace(st.Value); v != "" {
				return v
			}
		}
	}
	return fallback
}

func (s SettingsServices) boolValue(key string, fallback bool) bool {
	v := s.value(key, strconv.FormatBool(fallback))
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}

func (s SettingsServices) intValue(key string, fallback int) int {
	v := s.value(key, strconv.Itoa(fallback))
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func (s SettingsServices) listValue(key, fallback string) []string {
	raw := s.value(key, fallback)
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if v := strings.TrimSpace(p); v != "" {
			out = append(out, v)
		}
	}
	return out
}

// ChatEnabled reports whether chat is on.
func (s SettingsServices) ChatEnabled() bool {
	return s.boolValue(constants.SupportSettingChatEnabled, global.GVA_CONFIG.Support.Chat.Enabled)
}

// WelcomeMessage returns the configured welcome text.
func (s SettingsServices) WelcomeMessage() string {
	return s.value(constants.SupportSettingWelcomeMessage, global.GVA_CONFIG.Support.Chat.WelcomeMessage)
}

// MaxMessageBytes returns the sanitized-body byte cap.
func (s SettingsServices) MaxMessageBytes() int {
	n := s.intValue(constants.SupportSettingMaxMessageBytes, global.GVA_CONFIG.Support.Chat.MaxMessageBytes)
	if n <= 0 {
		return 4000
	}
	return n
}

// PageSize returns the default message page size.
func (s SettingsServices) PageSize() int {
	n := global.GVA_CONFIG.Support.Chat.PageSize
	if n <= 0 {
		return 50
	}
	if n > 100 {
		return 100
	}
	return n
}

// RequireFaqBeforeChat reports whether the FAQ entry is required before chatting.
func (s SettingsServices) RequireFaqBeforeChat() bool {
	return s.boolValue(constants.SupportSettingRequireFaqBeforeChat, global.GVA_CONFIG.Support.Chat.RequireFaqBeforeChat)
}

// QualityCanReadAll reports whether quality reviewers may read every conversation.
func (s SettingsServices) QualityCanReadAll() bool {
	return global.GVA_CONFIG.Support.Chat.QualityCanReadAll
}

// RateLimitPer10s returns the per-10s message budget.
func (s SettingsServices) RateLimitPer10s() int {
	return s.intValue(constants.SupportSettingRateLimitPer10s, global.GVA_CONFIG.Support.Chat.RateLimit.Per10s)
}

// RateLimitPerHour returns the hourly message budget.
func (s SettingsServices) RateLimitPerHour() int {
	return s.intValue(constants.SupportSettingRateLimitPerHour, global.GVA_CONFIG.Support.Chat.RateLimit.PerHour)
}

// AttachmentsEnabled reports whether uploads are allowed.
func (s SettingsServices) AttachmentsEnabled() bool {
	return s.boolValue(constants.SupportSettingAttachmentsEnabled, global.GVA_CONFIG.Support.Chat.Attachments.Enabled)
}

// MaxFileBytes returns the per-file cap, never above the storage layer's hard cap.
func (s SettingsServices) MaxFileBytes() int64 {
	n := int64(s.intValue(constants.SupportSettingAttachmentsMaxFileBytes, int(global.GVA_CONFIG.Support.Chat.Attachments.MaxFileBytes)))
	if n <= 0 || n > storageMaxFileBytes {
		return storageMaxFileBytes
	}
	return n
}

// MaxFilesPerMessage returns how many attachments one message may carry.
func (s SettingsServices) MaxFilesPerMessage() int {
	n := s.intValue(constants.SupportSettingAttachmentsMaxFiles, global.GVA_CONFIG.Support.Chat.Attachments.MaxFilesPerMessage)
	if n <= 0 {
		return 5
	}
	return n
}

// MaxTotalBytesPerMessage returns the summed attachment cap for one message.
func (s SettingsServices) MaxTotalBytesPerMessage() int64 {
	n := int64(s.intValue(constants.SupportSettingAttachmentsMaxTotal, int(global.GVA_CONFIG.Support.Chat.Attachments.MaxTotalBytesPerMessage)))
	if n <= 0 {
		return 20 << 20
	}
	return n
}

// AllowedMime returns the allowed MIME list, intersected with the storage layer's extension
// allowlist's MIME equivalents so an admin cannot widen it into a type that cannot be stored.
func (s SettingsServices) AllowedMime() []string {
	fallback := strings.Join(global.GVA_CONFIG.Support.Chat.Attachments.AllowedMime, ",")
	configured := s.listValue(constants.SupportSettingAttachmentsAllowedMime, fallback)
	storageMime := map[string]bool{
		"image/jpeg": true, "image/png": true, "image/gif": true, "image/webp": true,
		"application/pdf": true, "application/msword": true,
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
	}
	out := make([]string, 0, len(configured))
	for _, m := range configured {
		if storageMime[m] {
			out = append(out, m)
		}
	}
	return out
}

// LastSeenDisplay reports whether last-seen values are shown at all.
func (s SettingsServices) LastSeenDisplay() bool {
	return s.boolValue(constants.SupportSettingLastSeenDisplay, global.GVA_CONFIG.Support.Presence.LastSeenDisplay)
}

// HiddenLastSeenRoles returns the actor types whose last-seen value is suppressed server-side.
func (s SettingsServices) HiddenLastSeenRoles() []string {
	return s.listValue(constants.SupportSettingHideLastSeenForRoles, strings.Join(global.GVA_CONFIG.Support.Presence.HideLastSeenForRoles, ","))
}

// editableKeys is the allowlist of runtime-settable support settings. Nothing that names a
// storage backend, path or oss-type is ever editable here.
var editableKeys = map[string]bool{
	constants.SupportSettingChatEnabled:             true,
	constants.SupportSettingWelcomeMessage:          true,
	constants.SupportSettingWelcomeMessageAr:        true,
	constants.SupportSettingAllowedSenderTypes:      true,
	constants.SupportSettingMaxMessageBytes:         true,
	constants.SupportSettingRateLimitPer10s:         true,
	constants.SupportSettingRateLimitPerHour:        true,
	constants.SupportSettingEscalateAfterMinutes:    true,
	constants.SupportSettingRequireFaqBeforeChat:    true,
	constants.SupportSettingAttachmentsEnabled:      true,
	constants.SupportSettingAttachmentsMaxFileBytes: true,
	constants.SupportSettingAttachmentsMaxFiles:     true,
	constants.SupportSettingAttachmentsMaxTotal:     true,
	constants.SupportSettingAttachmentsAllowedMime:  true,
	constants.SupportSettingLastSeenDisplay:         true,
	constants.SupportSettingHideLastSeenForRoles:    true,
	constants.SupportSettingPushEnabled:             true,
}

// List returns every support setting with its effective value.
func (s SettingsServices) List() map[string]string {
	out := map[string]string{}
	for k := range editableKeys {
		out[k] = s.value(k, "")
	}
	return out
}

// Update writes runtime settings through the existing audited settings service.
func (s SettingsServices) Update(adminID uint, req supportReq.ChatSettingsUpdateReq) error {
	if len(req.Values) == 0 {
		return errors.New("no values supplied")
	}
	limit := global.GVA_CONFIG.Support.Admin.MaxWelcomeMessageBytes
	for key, value := range req.Values {
		if !editableKeys[key] {
			return errors.New("setting not editable: " + key)
		}
		if (key == constants.SupportSettingWelcomeMessage || key == constants.SupportSettingWelcomeMessageAr) &&
			limit > 0 && len(value) > limit {
			return errors.New("welcome message too long")
		}
		if key == constants.SupportSettingAttachmentsMaxFileBytes {
			n, err := strconv.ParseInt(value, 10, 64)
			if err != nil || n <= 0 {
				return errors.New("invalid attachment size limit")
			}
			if n > storageMaxFileBytes {
				return errors.New("attachment size limit exceeds the storage cap")
			}
		}
	}
	for key, value := range req.Values {
		if _, err := (base.LookupServices{}).UpdateSetting(&adminID, baseReq.SettingUpdateReq{
			Section: constants.SupportSettingSection, Key: key, Value: value, Description: req.Reason,
		}); err != nil {
			return err
		}
	}
	return nil
}
