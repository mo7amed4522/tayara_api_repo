package request

import (
	commonReq "github.com/mo7amed4522/tayara_api/model/common/request"
)

// CreateDisputeReq user/agent opens a dispute on a booking.
type CreateDisputeReq struct {
	BookingID uint   `json:"bookingId" binding:"required"`
	Details   string `json:"details" binding:"required"`
}

// ResolveDisputeReq admin resolves/closes a dispute.
type ResolveDisputeReq struct {
	DisputeID  uint   `json:"disputeId" binding:"required"`
	Status     string `json:"status"` // resolved | closed
	Resolution string `json:"resolution"`
}

// DisputePageReq paginated dispute listing with optional booking/status filter.
type DisputePageReq struct {
	commonReq.PageInfo
	BookingID uint   `json:"bookingId" form:"bookingId"`
	Status    string `json:"status" form:"status"`
}
