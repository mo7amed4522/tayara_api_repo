package constants

// Push dispatch statuses. Only "sent" and "collapsed" are terminal-delivered states; "failed" is
// retried by the sweeper until the attempt cap is reached.
const (
	PushStatusPending   = "pending"
	PushStatusSent      = "sent"
	PushStatusFailed    = "failed"
	PushStatusCollapsed = "collapsed"
)

// Push platforms accepted by the token registration endpoint.
const (
	PushPlatformIOS     = "ios"
	PushPlatformAndroid = "android"
	PushPlatformWeb     = "web"
)

// PushChannelFCM is the only delivery channel in this changeset.
const PushChannelFCM = "fcm"

// Runtime setting key for the push consumer (section `support`).
const SupportSettingPushEnabled = "support.push_enabled"

// SupportPushCollapseKeyPrefix namespaces the best-effort Redis collapse window.
const SupportPushCollapseKeyPrefix = "support:push:collapse:"

// Rate-limit buckets for the push token endpoints.
const (
	SupportRateLimitPushTokenRegister = "push-token-register"
	SupportRateLimitPushTokenRevoke   = "push-token-revoke"
)

// SupportPushNotificationEventType is the engage notification/log event type for support chat. It
// is also the key an admin notification template is matched on.
const SupportPushNotificationEventType = "support.chat.message"

// SupportPushNotificationSSEEvent is the real-time frame name the app listens to on
// /api/v1/support/notifications/stream.
const SupportPushNotificationSSEEvent = "notification.created"

// Rate-limit bucket for the real-time notification stream.
const SupportRateLimitNotificationStream = "notification-stream"
