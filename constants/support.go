package constants

// Support module enumerations. Every status is a lowercase snake string (never a numeric
// literal) so the database stays readable and the API contract is explicit.

// Support actor / sender types. Customers are bx_user_profile rows; everyone else is a
// sys_admin_profile row distinguished by the authority id.
const (
	ActorCustomer = "customer"
	ActorSupport  = "support"
	ActorQuality  = "quality"
	ActorAdmin    = "admin"
	ActorSystem   = "system"
)

// Support staff authority ids (stamped into the JWT at login and used by authorization checks).
const (
	AuthoritySupport = 4 // customer-support employee (sys_admin_profile)
	AuthorityQuality = 5 // quality reviewer (sys_admin_profile)
)

// Ticket lifecycle.
const (
	SupportTicketStatusOpen            = "open"
	SupportTicketStatusPendingCustomer = "pending_customer"
	SupportTicketStatusPendingSupport  = "pending_support"
	SupportTicketStatusSolved          = "solved"
	SupportTicketStatusNotSolved       = "not_solved"
	SupportTicketStatusClosed          = "closed"
)

// Ticket priority.
const (
	SupportPriorityLow    = "low"
	SupportPriorityNormal = "normal"
	SupportPriorityHigh   = "high"
	SupportPriorityUrgent = "urgent"
)

// Conversation lifecycle.
const (
	SupportConversationStatusActive = "active"
	SupportConversationStatusClosed = "closed"
)

// Quality review verdicts.
const (
	SupportVerdictSolved    = "solved"
	SupportVerdictNotSolved = "not_solved"
)

// FAQ vote options.
const (
	SupportVoteHelpful    = "helpful"
	SupportVoteNotHelpful = "not_helpful"
)

// Presence states. "unknown" is returned when Redis is unavailable — never an error.
const (
	SupportPresenceOnline  = "online"
	SupportPresenceOffline = "offline"
	SupportPresenceUnknown = "unknown"
)

// Attachment lifecycle. Exactly two transitions are legal, both from "uploaded".
const (
	SupportAttachmentStatusUploaded = "uploaded"
	SupportAttachmentStatusAttached = "attached"
	SupportAttachmentStatusRejected = "rejected"
)

// Attachment kinds, derived from the sniffed MIME type.
const (
	SupportAttachmentKindImage = "image"
	SupportAttachmentKindFile  = "file"
)

// Support ticket code prefix.
const CodePrefixSupportTicket = "ST-"

// Presence Redis key namespaces.
const (
	SupportPresenceConnKeyPrefix  = "support:presence:conn:"
	SupportPresenceIndexKey       = "support:presence:index"
	SupportPresenceLockPrefix     = "support:presence:"
	SupportConversationLockPrefix = "support:conv:"
)

// Runtime setting keys (base_system_setting, section `support`). DB value wins over YAML.
const (
	SupportSettingSection                 = "support"
	SupportSettingChatEnabled             = "support.chat_enabled"
	SupportSettingWelcomeMessage          = "support.welcome_message"
	SupportSettingWelcomeMessageAr        = "support.welcome_message_ar"
	SupportSettingAllowedSenderTypes      = "support.allowed_sender_types"
	SupportSettingMaxMessageBytes         = "support.max_message_bytes"
	SupportSettingRateLimitPer10s         = "support.rate_limit_per_10s"
	SupportSettingRateLimitPerHour        = "support.rate_limit_per_hour"
	SupportSettingEscalateAfterMinutes    = "support.escalate_after_minutes"
	SupportSettingRequireFaqBeforeChat    = "support.require_faq_before_chat"
	SupportSettingAttachmentsEnabled      = "support.attachments_enabled"
	SupportSettingAttachmentsMaxFileBytes = "support.attachments_max_file_bytes"
	SupportSettingAttachmentsMaxFiles     = "support.attachments_max_files_per_message"
	SupportSettingAttachmentsMaxTotal     = "support.attachments_max_total_bytes"
	SupportSettingAttachmentsAllowedMime  = "support.attachments_allowed_mime"
	SupportSettingLastSeenDisplay         = "support.last_seen_display"
	SupportSettingHideLastSeenForRoles    = "support.hide_last_seen_for_roles"
)

// Support settings section alias used by the base lookup service.
const SettingSectionSupport = SupportSettingSection
