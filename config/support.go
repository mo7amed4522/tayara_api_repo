package config

// SupportConfig is the root of the `support:` section: real-time support chat, presence,
// last seen and attachments. Kafka is optional — chat persists first and works without a broker.
type SupportConfig struct {
	Enabled  bool              `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
	Kafka    SupportKafka      `mapstructure:"kafka" json:"kafka" yaml:"kafka"`
	Chat     SupportChat       `mapstructure:"chat" json:"chat" yaml:"chat"`
	Presence SupportPresence   `mapstructure:"presence" json:"presence" yaml:"presence"`
	Admin    SupportAdminBlock `mapstructure:"admin" json:"admin" yaml:"admin"`
	Push     SupportPush       `mapstructure:"push" json:"push" yaml:"push"`
}

// SupportKafka is the broker transport configuration. Disabled by default so an existing
// deployment is unaffected.
type SupportKafka struct {
	Enabled           bool                 `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
	Brokers           []string             `mapstructure:"brokers" json:"brokers" yaml:"brokers"`
	ClientID          string               `mapstructure:"client-id" json:"client-id" yaml:"client-id"`
	RequiredAcks      string               `mapstructure:"required-acks" json:"required-acks" yaml:"required-acks"`
	Topics            SupportKafkaTopics   `mapstructure:"topics" json:"topics" yaml:"topics"`
	ConsumerGroup     string               `mapstructure:"consumer-group" json:"consumer-group" yaml:"consumer-group"`
	FanoutGroupPrefix string               `mapstructure:"fanout-group-prefix" json:"fanout-group-prefix" yaml:"fanout-group-prefix"`
	AutoCreateTopics  bool                 `mapstructure:"auto-create-topics" json:"auto-create-topics" yaml:"auto-create-topics"`
	Retry             SupportKafkaRetry    `mapstructure:"retry" json:"retry" yaml:"retry"`
	Producer          SupportKafkaProducer `mapstructure:"producer" json:"producer" yaml:"producer"`
	Consumer          SupportKafkaConsumer `mapstructure:"consumer" json:"consumer" yaml:"consumer"`
	TLS               SupportKafkaTLS      `mapstructure:"tls" json:"tls" yaml:"tls"`
	SASL              SupportKafkaSASL     `mapstructure:"sasl" json:"sasl" yaml:"sasl"`
}

type SupportKafkaTopics struct {
	ChatMessages    string `mapstructure:"chat-messages" json:"chat-messages" yaml:"chat-messages"`
	ChatEvents      string `mapstructure:"chat-events" json:"chat-events" yaml:"chat-events"`
	ChatMessagesDLQ string `mapstructure:"chat-messages-dlq" json:"chat-messages-dlq" yaml:"chat-messages-dlq"`
	Presence        string `mapstructure:"presence" json:"presence" yaml:"presence"`
}

type SupportKafkaRetry struct {
	MaxAttempts      int `mapstructure:"max-attempts" json:"max-attempts" yaml:"max-attempts"`
	InitialBackoffMS int `mapstructure:"initial-backoff-ms" json:"initial-backoff-ms" yaml:"initial-backoff-ms"`
	MaxBackoffMS     int `mapstructure:"max-backoff-ms" json:"max-backoff-ms" yaml:"max-backoff-ms"`
}

type SupportKafkaProducer struct {
	LingerMS   int `mapstructure:"linger-ms" json:"linger-ms" yaml:"linger-ms"`
	BatchBytes int `mapstructure:"batch-bytes" json:"batch-bytes" yaml:"batch-bytes"`
	TimeoutMS  int `mapstructure:"timeout-ms" json:"timeout-ms" yaml:"timeout-ms"`
}

type SupportKafkaConsumer struct {
	SessionTimeoutMS int  `mapstructure:"session-timeout-ms" json:"session-timeout-ms" yaml:"session-timeout-ms"`
	MaxPollRecords   int  `mapstructure:"max-poll-records" json:"max-poll-records" yaml:"max-poll-records"`
	StartFromOldest  bool `mapstructure:"start-from-oldest" json:"start-from-oldest" yaml:"start-from-oldest"`
}

type SupportKafkaTLS struct {
	Enabled            bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
	CAFile             string `mapstructure:"ca-file" json:"ca-file" yaml:"ca-file"`
	CertFile           string `mapstructure:"cert-file" json:"cert-file" yaml:"cert-file"`
	KeyFile            string `mapstructure:"key-file" json:"key-file" yaml:"key-file"`
	InsecureSkipVerify bool   `mapstructure:"insecure-skip-verify" json:"insecure-skip-verify" yaml:"insecure-skip-verify"`
}

type SupportKafkaSASL struct {
	Enabled   bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
	Mechanism string `mapstructure:"mechanism" json:"mechanism" yaml:"mechanism"`
	Username  string `mapstructure:"username" json:"username" yaml:"username"`
	Password  string `mapstructure:"password" json:"password" yaml:"password"`
}

type SupportChat struct {
	Enabled              bool                 `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
	WelcomeMessage       string               `mapstructure:"welcome-message" json:"welcome-message" yaml:"welcome-message"`
	WelcomeMessageAr     string               `mapstructure:"welcome-message-ar" json:"welcome-message-ar" yaml:"welcome-message-ar"`
	MaxMessageBytes      int                  `mapstructure:"max-message-bytes" json:"max-message-bytes" yaml:"max-message-bytes"`
	PageSize             int                  `mapstructure:"page-size" json:"page-size" yaml:"page-size"`
	HashChainEnabled     bool                 `mapstructure:"hash-chain-enabled" json:"hash-chain-enabled" yaml:"hash-chain-enabled"`
	RequireFaqBeforeChat bool                 `mapstructure:"require-faq-before-chat" json:"require-faq-before-chat" yaml:"require-faq-before-chat"`
	QualityCanReadAll    bool                 `mapstructure:"quality-can-read-all" json:"quality-can-read-all" yaml:"quality-can-read-all"`
	VoteChangePolicy     string               `mapstructure:"vote-change-policy" json:"vote-change-policy" yaml:"vote-change-policy"`
	RateLimit            SupportChatRateLimit `mapstructure:"rate-limit" json:"rate-limit" yaml:"rate-limit"`
	AllowedSenderTypes   []string             `mapstructure:"allowed-sender-types" json:"allowed-sender-types" yaml:"allowed-sender-types"`
	Assignment           SupportAssignment    `mapstructure:"assignment" json:"assignment" yaml:"assignment"`
	Attachments          SupportAttachments   `mapstructure:"attachments" json:"attachments" yaml:"attachments"`
}

type SupportChatRateLimit struct {
	Per10s  int `mapstructure:"per-10s" json:"per-10s" yaml:"per-10s"`
	PerHour int `mapstructure:"per-hour" json:"per-hour" yaml:"per-hour"`
}

type SupportAssignment struct {
	AllowRedirect                  bool `mapstructure:"allow-redirect" json:"allow-redirect" yaml:"allow-redirect"`
	RequireReason                  bool `mapstructure:"require-reason" json:"require-reason" yaml:"require-reason"`
	EscalateAfterMinutesUnanswered int  `mapstructure:"escalate-after-minutes-unanswered" json:"escalate-after-minutes-unanswered" yaml:"escalate-after-minutes-unanswered"`
}

type SupportAttachments struct {
	Enabled                 bool     `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
	MaxFileBytes            int64    `mapstructure:"max-file-bytes" json:"max-file-bytes" yaml:"max-file-bytes"`
	MaxFilesPerMessage      int      `mapstructure:"max-files-per-message" json:"max-files-per-message" yaml:"max-files-per-message"`
	MaxTotalBytesPerMessage int64    `mapstructure:"max-total-bytes-per-message" json:"max-total-bytes-per-message" yaml:"max-total-bytes-per-message"`
	AllowedMime             []string `mapstructure:"allowed-mime" json:"allowed-mime" yaml:"allowed-mime"`
	OrphanCleanupHours      int      `mapstructure:"orphan-cleanup-hours" json:"orphan-cleanup-hours" yaml:"orphan-cleanup-hours"`
	ThumbnailWidth          int      `mapstructure:"thumbnail-width" json:"thumbnail-width" yaml:"thumbnail-width"`
}

type SupportPresence struct {
	Enabled                  bool     `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
	HeartbeatIntervalSeconds int      `mapstructure:"heartbeat-interval-seconds" json:"heartbeat-interval-seconds" yaml:"heartbeat-interval-seconds"`
	TTLSeconds               int      `mapstructure:"ttl-seconds" json:"ttl-seconds" yaml:"ttl-seconds"`
	AwayAfterSeconds         int      `mapstructure:"away-after-seconds" json:"away-after-seconds" yaml:"away-after-seconds"`
	LastSeenFlushSeconds     int      `mapstructure:"last-seen-flush-seconds" json:"last-seen-flush-seconds" yaml:"last-seen-flush-seconds"`
	LastSeenDisplay          bool     `mapstructure:"last-seen-display" json:"last-seen-display" yaml:"last-seen-display"`
	HideLastSeenForRoles     []string `mapstructure:"hide-last-seen-for-roles" json:"hide-last-seen-for-roles" yaml:"hide-last-seen-for-roles"`
	BatchReadMaxIds          int      `mapstructure:"batch-read-max-ids" json:"batch-read-max-ids" yaml:"batch-read-max-ids"`
}

type SupportAdminBlock struct {
	SettingsSection           string `mapstructure:"settings-section" json:"settings-section" yaml:"settings-section"`
	AllowWelcomeTextEdit      bool   `mapstructure:"allow-welcome-text-edit" json:"allow-welcome-text-edit" yaml:"allow-welcome-text-edit"`
	AllowChatToggle           bool   `mapstructure:"allow-chat-toggle" json:"allow-chat-toggle" yaml:"allow-chat-toggle"`
	AllowAllowedRolesEdit     bool   `mapstructure:"allow-allowed-roles-edit" json:"allow-allowed-roles-edit" yaml:"allow-allowed-roles-edit"`
	AllowAttachmentLimitsEdit bool   `mapstructure:"allow-attachment-limits-edit" json:"allow-attachment-limits-edit" yaml:"allow-attachment-limits-edit"`
	RequireChangeReason       bool   `mapstructure:"require-change-reason" json:"require-change-reason" yaml:"require-change-reason"`
	MaxWelcomeMessageBytes    int    `mapstructure:"max-welcome-message-bytes" json:"max-welcome-message-bytes" yaml:"max-welcome-message-bytes"`
}
