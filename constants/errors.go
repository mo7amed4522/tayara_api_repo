package constants

// API error codes (normalized across the platform). The response envelope is always
// `{code, data, msg}`; the HTTP status mirrors the code's class (4xx client, 5xx server)
// while the `code` field carries the machine-readable id the dashboard can switch on.
// Keep this file the single source of truth for error semantics.
const (
	// CodeSuccess 0 — request succeeded (see model/common/response.SUCCESS).
	CodeSuccess = 0

	// CodeUnauthorized 401 → credential missing/expired/invalid (HTTP 401).
	CodeUnauthorized = 401
	// CodeForbidden 403 — authenticated but not allowed (HTTP 403).
	CodeForbidden = 403
	// CodeNotFound 404 — resource not found (HTTP 404).
	CodeNotFound = 404
	// CodeConflict 409 — state conflict, e.g. an illegal lifecycle transition or a
	// duplicate that must not exist (HTTP 409).
	CodeConflict = 409
	// CodeValidation 422 — request body/query failed validation (HTTP 422).
	CodeValidation = 422

	// CodeInternal 500 — unexpected server error (HTTP 500).
	CodeInternal = 500
	// CodeDB 5001 — persistence error (HTTP 500).
	CodeDB = 5001
	// CodeProvider 502 — third-party provider (Stripe/BNPL/email/fcm) failure (HTTP 502).
	CodeProvider = 502
)
