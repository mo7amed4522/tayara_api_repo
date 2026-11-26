package support

// ServiceGroup groups the support services; embedded into service.ServiceGroupApp.
type ServiceGroup struct {
	TicketServices
	MessageServices
	AuthorizationServices
	SettingsServices
	PresenceServices
	LastSeenServices
	AttachmentServices
	AssignmentServices
	ReviewServices
	FaqServices
	OutboxServices
}
