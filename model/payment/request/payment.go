package request

// CreatePaymentReq opens a payment for a booking with a chosen method.
type CreatePaymentReq struct {
	BookingID     uint   `json:"bookingId" binding:"required"`
	PaymentMethod string `json:"paymentMethod" binding:"required"` // card|apple_pay|google_pay|tabby|tamara
}

// CaptureReq captures an authorized payment for a booking (admin/system).
type CaptureReq struct {
	BookingID uint `json:"bookingId" binding:"required"`
}

// RefundReq refunds a captured payment (admin, post-capture).
type RefundReq struct {
	BookingID  uint   `json:"bookingId" binding:"required"`
	RefundType string `json:"refundType" binding:"required"` // full|partial
	Amount     int64  `json:"amount"`                        // minor units; required (<= captured) for partial
	Reason     string `json:"reason" binding:"required"`
}
