package constants

// Marketplace enums (Phase 0). Companion to reason.go — these cover the travel-marketplace domain
// that reason.go does not already define. Single source of truth for their values.

// KYC status axis (user identity verification, distinct from per-traveler DocumentStatus)
const (
	KYCStatusNotSubmitted = "not_submitted"
	KYCStatusPending      = "pending"
	KYCStatusApproved     = "approved"
	KYCStatusRejected     = "rejected"
)

// Document role within a booking (booking_documents.doc_role)
const (
	DocRoleKYCFile           = "kyc_file"
	DocRoleAgentConfirmation = "agent_confirmation"
	DocRoleTravelerDoc       = "traveler_doc"
	DocRoleIssuedTicket      = "issued_ticket"
)

// Document requirement phase (package_document_requirements.required_after)
const (
	RequireAfterBooking = "booking"
	RequireAfterPayment = "payment"
)

// Package availability_type
const (
	AvailabilityOpenCalendar = "open_calendar"
	AvailabilityFixedDates   = "fixed_dates"
)

// Package add-on types + pricing model
const (
	AddonTypeVisaIssueHelp   = "visa_issue_help"
	AddonTypeTourGuide       = "tour_guide"
	AddonTypeTravelInsurance = "travel_insurance"
	AddonTypePhotographer    = "photographer"
)
const (
	AddonPriceModelFixed       = "fixed"        // tour guide / photographer
	AddonPriceModelPerTraveler = "per_traveler" // visa help / insurance
)

// Ad placement and redirect targets
const (
	AdPlacementHomeTopBanner    = "home_top_banner"
	AdPlacementExploreTopBanner = "explore_top_banner"
)
const (
	AdRedirectPackagesByDestination = "packages_by_destination"
	AdRedirectAgentsByDestination   = "agents_by_destination"
	AdRedirectSingleAgentPackages   = "single_agent_packages"
)

// Notification event types (notifications.event_type / notification_templates)
const (
	NotifEventBookingRequested = "booking_request"
	NotifEventAgentReviewing   = "agent_reviewing"
	NotifEventPriceConfirmed   = "price_confirmed"
	NotifEventOfferExpiring    = "offer_expiring"
	NotifEventAgentRejected    = "agent_rejected"
	NotifEventPaymentSuccess   = "payment_success"
	NotifEventSLATimeout       = "sla_timeout"
	NotifEventCancellation     = "cancellation"
	NotifEventPayout           = "payout"
	NotifEventRefund           = "refund"
)

// Notification recipient type
const (
	NotifRecipientUser  = "user"
	NotifRecipientAgent = "agent"
	NotifRecipientAdmin = "admin"
)

// Notification channel
const (
	NotifChannelPush  = "push"
	NotifChannelEmail = "email"
)

// SLA enforcement types (sla_events.sla_type)
const (
	SLATypeAgentResponse = "agent_response" // agent must respond to a request within SLA hours
	SLATypeUserPayment   = "user_payment"   // user must pay within the pay window after price confirmation
)

// SLA auto-action taken on breach
const (
	SLAAutoActionAutoCancel  = "auto_cancel"
	SLAAutoActionOfferExpiry = "offer_expiry"
	SLAAutoActionReminder    = "reminder"
)

// Currency codes (default AED; supported AED/USD)
const (
	CurrencyAED = "AED"
	CurrencyUSD = "USD"
)

// Payment status also includes a pre-authorization pending state (added here; see reason.go for the rest).
const (
	PaymentStatusPending = "pending"
)

// Payment providers and methods (Stripe cards/wallets + Tabby/Tamara BNPL).
const (
	PaymentProviderStripe = "stripe"
	PaymentProviderTabby  = "tabby"
	PaymentProviderTamara = "tamara"
)
const (
	PaymentMethodCard      = "card"
	PaymentMethodApplePay  = "apple_pay"
	PaymentMethodGooglePay = "google_pay"
	PaymentMethodTabby     = "tabby"
	PaymentMethodTamara    = "tamara"
)

// Payment lifecycle statuses (Mongo state + MySQL payment_transactions.status).
const (
	PaymentLifecycleCreated           = "created"
	PaymentLifecycleAuthorized        = "authorized"
	PaymentLifecycleCaptured          = "captured"
	PaymentLifecyclePartiallyRefunded = "partially_refunded"
	PaymentLifecycleRefunded          = "refunded"
	PaymentLifecycleFailed            = "failed"
)

// Payment lifecycle event types recorded in the Mongo ledger.
const (
	PaymentEventCreated    = "created"
	PaymentEventAuthorized = "authorized"
	PaymentEventCaptured   = "captured"
	PaymentEventRefunded   = "refunded"
	PaymentEventFailed     = "failed"
)

// Coupon lifecycle statuses (display-scope MVP).
const (
	CouponStatusActive  = "active"
	CouponStatusUsed    = "used"
	CouponStatusExpired = "expired"
)

// RBAC authority ids carried in the JWT claims (AuthorityId). These map to the domain guard
// groups (system/admin, manage/agent, bx/user) and drive route-authority checks centrally.
const (
	AuthoritySystem = 1 // platform admin (super admin)
	AuthorityAgent  = 2 // travel agency dashboard
	AuthorityUser   = 3 // end-user mobile app
)

// Dispute lifecycle statuses (booking disputes).
const (
	DisputeStatusOpen     = "open"
	DisputeStatusResolved = "resolved"
	DisputeStatusClosed   = "closed"
)

// Commission record plan-type snapshot (commission_records.plan_type).
const (
	CommissionPlanBasic   = "basic"
	CommissionPlanPremium = "premium"
	CommissionPlanCustom  = "custom"
)

// Commission record lifecycle status.
const (
	CommissionStatusCaptured = "captured"
	CommissionStatusPaid     = "paid"
)

// Requirement profile phases for custom-trip documents.
const (
	RequireProfileBefore = "before"
	RequireProfileAfter  = "after"
)

// Discount floor (PRD locks a minimum discount on packages). Basis points, 300 == 3%.
const (
	DiscountMinPercentBps = 300
)

// Code prefixes used by the utils code generator (PRD-named sequence prefixes)
const (
	CodePrefixBooking      = "BK-"
	CodePrefixCustomTrip   = "CT-"
	CodePrefixPackage      = "PK-"
	CodePrefixAgent        = "AG-"
	CodePrefixUser         = "US-"
	CodePrefixNotification = "NT-"
	CodePrefixAd           = "AD-"
	CodePrefixPayment      = "PAY-"
)
