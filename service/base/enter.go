package base

// BaseServices groups the base/dictionaries services (lookups + audit history).
type BaseServices struct {
	LookupServices
}

// BaseServiceApp default instance, matched to the `service/<module>/enter.go` convention.
var BaseServiceApp = new(BaseServices)
