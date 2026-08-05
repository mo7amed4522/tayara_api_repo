package request

import (
	commonReq "github.com/mo7amed4522/tayara_api/model/common/request"
)

// AgentRegisterReq public registration payload for a travel agency.
type AgentRegisterReq struct {
	CompanyName      string `json:"companyName" binding:"required"`
	ResponsibleName  string `json:"responsibleName" binding:"required"`
	BusinessEmail    string `json:"businessEmail" binding:"required,email"`
	Password         string `json:"password" binding:"required,min=6"`
	PhoneCountryCode string `json:"phoneCountryCode"`
	Phone            string `json:"phone"`
	Country          string `json:"country"`
	CompanyAddress   string `json:"companyAddress"`
	LicenseNumber    string `json:"licenseNumber"`
	PlanID           *uint  `json:"planId"`
}

// AgentLoginReq agent login credentials.
type AgentLoginReq struct {
	BusinessEmail string `json:"businessEmail" binding:"required,email"`
	Password      string `json:"password" binding:"required"`
}

// AgentApproveReq admin approves an agent and sets commission % (bps) + SLA hours at approval.
type AgentApproveReq struct {
	AgentID           uint  `json:"agentId" binding:"required"`
	CommissionPercent int64 `json:"commissionPercent"` // basis points; 0 = default from plan
	SLAHours          int   `json:"slaHours"`          // 0 = keep default
}

// AgentRejectReq admin rejects an agent with a reason.
type AgentRejectReq struct {
	AgentID uint   `json:"agentId" binding:"required"`
	Reason  string `json:"reason"`
}

// AgentSuspendReq admin suspends/reactivates an agent.
type AgentSuspendReq struct {
	AgentID uint   `json:"agentId" binding:"required"`
	Reason  string `json:"reason"`
}

// AgentNoteReq admin adds an internal note on an agent.
type AgentNoteReq struct {
	AgentID  uint   `json:"agentId" binding:"required"`
	NoteText string `json:"noteText" binding:"required"`
}

// AgentPageReq paginated agent listing (admin) with optional status filter.
type AgentPageReq struct {
	commonReq.PageInfo
	Status string `json:"status" form:"status"`
}

// ChangeAgentPasswordReq agent changes their own password.
type ChangeAgentPasswordReq struct {
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required,min=6"`
}
