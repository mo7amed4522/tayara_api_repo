package support

import "github.com/mo7amed4522/tayara_api/service/support/push"

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
	push.ServiceGroup
}
