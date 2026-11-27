package constants

// Recommendation system enums (stored recommendations, audit runs and the public list).
// Single source of truth for the values persisted in recommendation / recommendation_run.

// Recommendation target types (recommendation.target_type)
const (
	RecommendationTargetTypePackage = "package" // catalog package
	RecommendationTargetTypeTrip    = "trip"    // custom-trip template
	RecommendationTargetTypeAgency  = "agency"  // travel agency (agent profile)
)

// Recommendation review status (recommendation.status)
const (
	RecommendationStatusPending  = "pending"  // waiting for an admin decision
	RecommendationStatusApproved = "approved" // served to customers
	RecommendationStatusRejected = "rejected" // hidden by an admin
	RecommendationStatusExpired  = "expired"  // no longer returned by the latest run
)

// Recommendation source (recommendation.source)
const (
	RecommendationSourceLLM   = "llm"   // produced by the LLM
	RecommendationSourceRules = "rules" // produced by the rules fallback
	RecommendationSourceAdmin = "admin" // created or curated by an admin
)

// Recommendation run status (recommendation_run.status)
const (
	RecommendationRunStatusRunning  = "running"  // the run is in progress
	RecommendationRunStatusSuccess  = "success"  // the LLM answered and its output was validated
	RecommendationRunStatusFailed   = "failed"   // the run could not complete (no rows persisted)
	RecommendationRunStatusFallback = "fallback" // the LLM failed and the rules ranking was used instead
)

// Recommendation audience
const (
	RecommendationDefaultAudience = "all" // audience used when a caller does not specify one
)

// Rate-limit buckets for the public recommendation endpoints (middleware.SupportRateLimit).
const (
	RecommendationRateLimitEvent = "recommendation-event"
)
