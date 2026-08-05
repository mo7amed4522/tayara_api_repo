package request

import (
	commonReq "github.com/mo7amed4522/tayara_api/model/common/request"
)

// PayAgentReq admin confirms a payout for an eligible booking.
type PayAgentReq struct {
	BookingID       uint   `json:"bookingId" binding:"required"`
	Method          string `json:"method"` // manual (MVP) | stripe_transfer
	ReferenceNumber string `json:"referenceNumber"`
	Notes           string `json:"notes"`
}

// PayoutPageReq paginated payout listing with optional status/agent filter.
type PayoutPageReq struct {
	commonReq.PageInfo
	Status  string `json:"status" form:"status"`
	AgentID uint   `json:"agentId" form:"agentId"`
}
