package constants

const (
	A = 1  //Login failed more than 3 times
	B = 2  //Reached the limit threshold
	C = 3  //Continuous n orders with utr but not successful
	D = 4  //Continuous n orders failed for the payment method
	E = 5  //Reached the temporary limit of the payment method
	F = 6  //The payment method is closed, and the order with the same utr is captured (rebound)
	G = 7  //The payment method is opened, and the order with more than 10 yuan is captured
	Z = -1 //Unknown
)

const (
	RoleSuperAdmin = "super_admin"
	RoleAdmin      = "admin"
	RoleAgent      = "agent"
	RoleClient     = "client"
)

const (
	UserStatusPending  = "pending"
	UserStatusActive   = "active"
	UserStatusBlocked  = "blocked"
	UserStatusDisabled = "disabled"
)

const (
	AuthProviderEmail    = "email"
	AuthProviderGoogle   = "google"
	AuthProviderApple    = "apple"
	AuthProviderFacebook = "facebook"
	AuthProviderPhone    = "phone"
)

const (
	AgentStatusPending   = "pending"
	AgentStatusActive    = "active"
	AgentStatusRejected  = "rejected"
	AgentStatusSuspended = "suspended"
)

const (
	PackageStatusDraft           = "draft"
	PackageStatusPendingApproval = "pending_approval"
	PackageStatusActive          = "active"
	PackageStatusDisabled        = "disabled"
	PackageStatusFlagged         = "flagged"
	PackageStatusExpired         = "expired"
)

const (
	BookingStatusRequested             = "requested"
	BookingStatusAgentReviewing        = "agent_reviewing"
	BookingStatusPriceConfirmed        = "price_confirmed"
	BookingStatusPriceExpired          = "price_expired"
	BookingStatusAwaitingPayment       = "awaiting_payment"
	BookingStatusPaymentAuthorized     = "payment_authorized"
	BookingStatusPaidConfirmed         = "paid_confirmed"
	BookingStatusDocsPending           = "docs_pending"
	BookingStatusDocsSubmitted         = "docs_submitted"
	BookingStatusKycPending            = "kyc_pending"
	BookingStatusKycApproved           = "kyc_approved"
	BookingStatusDocsPendingCorrection = "docs_pending_correction"
	BookingStatusBookingConfirmed      = "booking_confirmed"
	BookingStatusProcessing            = "processing"
	BookingStatusIssuedReady           = "issued_ready"
	BookingStatusCompleted             = "completed"
	BookingStatusCancelled             = "cancelled"
	BookingStatusRefunded              = "refunded"
)

const (
	OfferStatusActive    = "active"
	OfferStatusExpired   = "expired"
	OfferStatusAccepted  = "accepted"
	OfferStatusCancelled = "cancelled"
)

const (
	PaymentStatusAuthorized        = "authorized"
	PaymentStatusCaptured          = "captured"
	PaymentStatusCancelled         = "cancelled"
	PaymentStatusRefunded          = "refunded"
	PaymentStatusPartiallyRefunded = "partially_refunded"
	PaymentStatusFailed            = "failed"
)

const (
	PayoutStatusNotEligible = "not_eligible"
	PayoutStatusPending     = "pending"
	PayoutStatusPaid        = "paid"
	PayoutStatusFailed      = "failed"
)

const (
	DocumentStatusPending           = "pending"
	DocumentStatusSubmitted         = "submitted"
	DocumentStatusApproved          = "approved"
	DocumentStatusRejected          = "rejected"
	DocumentStatusPendingCorrection = "pending_correction"
)

const (
	AdStatusDraft     = "draft"
	AdStatusScheduled = "scheduled"
	AdStatusActive    = "active"
	AdStatusDisabled  = "disabled"
	AdStatusExpired   = "expired"
)

const (
	NotificationStatusPending = "pending"
	NotificationStatusSent    = "sent"
	NotificationStatusFailed  = "failed"
	NotificationStatusRead    = "read"
)

const (
	RefundStatusPending  = "pending"
	RefundStatusRefunded = "refunded"
	RefundStatusFailed   = "failed"
)

const (
	RefundTypeFull    = "full"
	RefundTypePartial = "partial"
)

const (
	PackageFlagStatusOpen     = "open"
	PackageFlagStatusResolved = "resolved"
)

const (
	ReviewStatusActive   = "active"
	ReviewStatusDisabled = "disabled"
)

const (
	AuditActionAgentApproval  = "agent_approval"
	AuditActionStatusChange   = "status_change"
	AuditActionPaymentCapture = "payment_capture"
	AuditActionRefundCreate   = "refund_create"
	AuditActionPayoutRelease  = "payout_release"
	AuditActionSettingsUpdate = "settings_update"
	AuditActionPriceChange    = "price_change"
	AuditActionKycReview      = "kyc_review"
	AuditActionDocumentAccess = "document_access"
	AuditActionPackageSubmit  = "package_submit"
	AuditActionPackageDisable = "package_disable"
	AuditActionDisputeResolve = "dispute_resolve"
)

const (
	AuditEntityTypePayment    = "payment"
	AuditEntityTypeBooking    = "booking"
	AuditEntityTypeAgent      = "agent"
	AuditEntityTypePackage    = "package"
	AuditEntityTypeSettings   = "settings"
	AuditEntityTypeAd         = "ad"
	AuditEntityTypePayout     = "payout"
	AuditEntityTypeUser       = "user"
	AuditEntityTypeCommission = "commission"
)

// Booking types
const (
	BookingTypePackage    = "package"
	BookingTypeCustomTrip = "custom_trip"
)

// Package types
const (
	PackageTypeInbound  = "inbound"
	PackageTypeOutbound = "outbound"
)

// Custom trip request statuses
const (
	CustomTripStatusPending        = "pending"
	CustomTripStatusReviewing      = "reviewing"
	CustomTripStatusOfferSubmitted = "offer_submitted"
	CustomTripStatusRejected       = "rejected"
)

// Payout method constants
const (
	PayoutMethodManual         = "manual"
	PayoutMethodStripeTransfer = "stripe_transfer"
)

// Timeline actor constants
const (
	TimelineActorAdmin  = "admin"
	TimelineActorSystem = "system"
	TimelineActorAgent  = "agent"
	TimelineActorUser   = "user"
)

// Timeline action constants
const (
	TimelineActionAdminOverrideStatus = "admin_override_status"
	TimelineActionAdminCapturePayment = "admin_capture_payment"
	TimelineActionAdminRefundPayment  = "admin_refund_payment"
	TimelineActionAdminApproveKYC     = "admin_approve_kyc"
	TimelineActionAdminRejectKYC      = "admin_reject_kyc"
	TimelineActionAdminForceCancel    = "admin_force_cancel"
	TimelineActionAdminReleasePayout  = "admin_release_payout"
	TimelineActionCommissionOverride  = "commission_override"
	TimelineActionPayoutReleased      = "payout_released"
)

// Notification event types
const (
	NotificationEventPayoutReleased  = "payout_released"
	NotificationEventKYCApproved     = "kyc_approved"
	NotificationEventKYCRejected     = "kyc_rejected"
	NotificationEventForceCancel     = "force_cancel"
	NotificationEventPackageApproved = "package_approved"
)

// Currency defaults
const (
	DefaultCurrency = "AED"
)

// Booking status added after payment capture
const (
	BookingStatusPaymentCaptured = "payment_captured"
)
