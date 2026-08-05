package booking

import "github.com/mo7amed4522/tayara_api/service"

// ApiGroup groups the booking api handlers; embedded into api.ApiGroupApp.BookingApiGroup.
type ApiGroup struct {
	BookingApi
	DocApi
	DisputeApi
}

// bookingService reaches the booking services via the group var.
var bookingService = service.ServiceGroupApp.BookingServiceGroup
