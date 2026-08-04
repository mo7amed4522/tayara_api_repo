package base

// BaseServices placeholder service instance for the base/dictionaries module.
// Business logic (read/update of system settings, lookups) lands here in later phases.
type BaseServices struct{}

// BaseServiceApp default instance, matched to the `service/<module>/enter.go` convention.
var BaseServiceApp = new(BaseServices)
