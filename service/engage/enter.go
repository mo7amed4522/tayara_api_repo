package engage

// ServiceGroup groups the engage services; embedded into service.ServiceGroupApp.EngageServiceGroup.
type ServiceGroup struct {
	NotificationServices
	ReviewServices
	AdServices
	UserEngageServices
}
