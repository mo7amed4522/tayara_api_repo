package config

// SupportPush is the offline delivery block. A dedicated Kafka consumer decides, per message and
// per recipient, whether the in-app SSE frame already reached the user (presence online) or a
// system push notification is required. Disabled by default so existing deployments are
// unaffected.
type SupportPush struct {
	Enabled               bool             `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
	CollapseWindowSeconds int              `mapstructure:"collapse-window-seconds" json:"collapse-window-seconds" yaml:"collapse-window-seconds"`
	MaxPreviewBytes       int              `mapstructure:"max-preview-bytes" json:"max-preview-bytes" yaml:"max-preview-bytes"`
	MaxTitleBytes         int              `mapstructure:"max-title-bytes" json:"max-title-bytes" yaml:"max-title-bytes"`
	SilentForRoles        []string         `mapstructure:"silent-for-roles" json:"silent-for-roles" yaml:"silent-for-roles"`
	Retry                 SupportPushRetry `mapstructure:"retry" json:"retry" yaml:"retry"`
	FCM                   SupportPushFCM   `mapstructure:"fcm" json:"fcm" yaml:"fcm"`
}

// SupportPushRetry mirrors the shape of support.kafka.retry so the same capped backoff values are
// reused; only the attempt cap is separate for push.
type SupportPushRetry struct {
	MaxAttempts int `mapstructure:"max-attempts" json:"max-attempts" yaml:"max-attempts"`
}

// SupportPushFCM carries the credentials reference. The Firebase SDK itself reads the service
// account from GOOGLE_APPLICATION_CREDENTIALS, so these keys are documentation/ops handles.
type SupportPushFCM struct {
	CredentialsFile string `mapstructure:"credentials-file" json:"credentials-file" yaml:"credentials-file"`
	ProjectID       string `mapstructure:"project-id" json:"project-id" yaml:"project-id"`
}
