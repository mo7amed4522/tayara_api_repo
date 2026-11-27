package service

import (
	"github.com/mo7amed4522/tayara_api/service/analytics"
	"github.com/mo7amed4522/tayara_api/service/booking"
	"github.com/mo7amed4522/tayara_api/service/bxapi"
	"github.com/mo7amed4522/tayara_api/service/catalog"
	"github.com/mo7amed4522/tayara_api/service/dxapi"
	"github.com/mo7amed4522/tayara_api/service/engage"
	"github.com/mo7amed4522/tayara_api/service/export"
	"github.com/mo7amed4522/tayara_api/service/payment"
	recommendationService "github.com/mo7amed4522/tayara_api/service/recommendation"
	"github.com/mo7amed4522/tayara_api/service/search"
	"github.com/mo7amed4522/tayara_api/service/support"
	"github.com/mo7amed4522/tayara_api/service/system"
	"github.com/mo7amed4522/tayara_api/service/ticket"
)

var ServiceGroupApp = new(ServiceGroup)

type ServiceGroup struct {
	SystemServiceGroup    system.ServiceGroup
	BxApiServiceGroup     bxapi.ServiceGroup
	ManageServiceGroup    dxapi.ServiceGroup
	CatalogServiceGroup   catalog.ServiceGroup
	BookingServiceGroup   booking.ServiceGroup
	EngageServiceGroup    engage.ServiceGroup
	PaymentServiceGroup   payment.ServiceGroup
	AnalyticsServiceGroup analytics.ServiceGroup
	SearchServiceGroup    search.ServiceGroup
	ExportServiceGroup    export.ServiceGroup
	TicketServiceGroup    ticket.ServiceGroup
	SupportServiceGroup   support.ServiceGroup
	RecommendationServiceGroup recommendationService.ServiceGroup
}
