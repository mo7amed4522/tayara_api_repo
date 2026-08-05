package booking

import api "github.com/mo7amed4522/tayara_api/api/v1"

// RouterGroup groups the booking routers; embedded into router.RouterGroupApp.Booking.
type RouterGroup struct {
	BookingRouter
}

var (
	bookingApi = api.ApiGroupApp.BookingApiGroup.BookingApi
	docApi     = api.ApiGroupApp.BookingApiGroup.DocApi
)
