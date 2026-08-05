package catalog

// ServiceGroup groups the catalog services; embedded into service.ServiceGroupApp.CatalogServiceGroup
// so api/v1/catalog reaches them via the group var.
type ServiceGroup struct {
	PackageServices
	CustomTripServices
}
