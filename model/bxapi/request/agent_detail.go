package request

import commonReq "github.com/mo7amed4522/tayara_api/model/common/request"

// CustomerAgentDetailReq params for the customer agent detail page.
type CustomerAgentDetailReq struct {
	commonReq.PageInfo
	// AgentID is the target agent (required).
	AgentID uint `json:"agentId" form:"agentId" binding:"required"`
	// DestinationCountryID filters packages to a specific destination; empty = all.
	DestinationCountryID *uint `json:"destinationCountryId" form:"destinationCountryId"`
}
