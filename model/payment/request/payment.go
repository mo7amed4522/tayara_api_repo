package request

import (
	commonReq "github.com/mo7amed4522/tayara_api/model/common/request"
)

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

// CommissionOverrideReq admin overrides commission % on a captured booking (reason required).
type CommissionOverrideReq struct {
	BookingID         uint  `json:"bookingId" binding:"required"`
	CommissionPercent int64 `json:"commissionPercent" binding:"required"`
	Reason            string `json:"reason" binding:"required"`
}

// CommissionPageReq paginated commission-record listing with optional booking/agent filter.
type CommissionPageReq struct {
	commonReq.PageInfo
	BookingID uint `json:"bookingId" form:"bookingId"`
	AgentID   uint `json:"agentId" form:"agentId"`
}
