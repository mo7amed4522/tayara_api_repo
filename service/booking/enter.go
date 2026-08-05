package booking

// ServiceGroup groups the booking services; embedded into service.ServiceGroupApp.BookingServiceGroup.
type ServiceGroup struct {
	BookingServices
	DocServices
	DisputeServices
}
