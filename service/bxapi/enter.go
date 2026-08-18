package bxapi

// ServiceGroup groups the bxapi services; embedded into service.ServiceGroupApp.BxApiServiceGroup.
type ServiceGroup struct {
	UserServices
	HomeServices
	AgentDetailServices
	PackageDetailServices
}
