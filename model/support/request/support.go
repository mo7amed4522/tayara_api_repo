package request

import "github.com/mo7amed4522/tayara_api/model/common"

// OpenTicketReq opens a support ticket. bookingId/disputeId are optional; a free-standing
// ticket is valid.
type OpenTicketReq struct {
	Subject   string `json:"subject" form:"subject" binding:"required,max=255"`
	Priority  string `json:"priority" form:"priority"`
	BookingID *uint  `json:"bookingId" form:"bookingId"`
	DisputeID *uint  `json:"disputeId" form:"disputeId"`
}

// SendMessageReq appends a message. Body may be empty only when at least one attachment is
// claimed. clientMsgId is the idempotency key.
type SendMessageReq struct {
	Body            string   `json:"body" form:"body"`
	ClientMsgID     string   `json:"clientMsgId" form:"clientMsgId" binding:"required,min=8,max=64"`
	AttachmentUuids []string `json:"attachmentUuids" form:"attachmentUuids"`
}

// AssignTicketReq reassigns a ticket to an employee.
type AssignTicketReq struct {
	EmployeeID  uint   `json:"employeeId" form:"employeeId" binding:"required"`
	AuthorityID uint   `json:"authorityId" form:"authorityId"` // 4=support (default), 5=quality
	Reason      string `json:"reason" form:"reason" binding:"max=500"`
}

// ReviewTicketReq records a quality verdict.
type ReviewTicketReq struct {
	Verdict string `json:"verdict" form:"verdict" binding:"required,oneof=solved not_solved"`
	Notes   string `json:"notes" form:"notes" binding:"max=2000"`
}

// VoteReq casts (or changes) a vote on a FAQ entry.
type VoteReq struct {
	OptionKey string `json:"optionKey" form:"optionKey" binding:"required,max=64"`
}

// FaqReq creates or updates a FAQ entry.
type FaqReq struct {
	ID         uint           `json:"id" form:"id"`
	Code       string         `json:"code" form:"code" binding:"required,max=64"`
	Question   string         `json:"question" form:"question" binding:"required,max=500"`
	QuestionAr string         `json:"questionAr" form:"questionAr"`
	Answer     string         `json:"answer" form:"answer" binding:"required"`
	AnswerAr   string         `json:"answerAr" form:"answerAr"`
	Options    common.JSONMap `json:"options" form:"options"`
	IsActive   bool           `json:"isActive" form:"isActive"`
	SortOrder  int            `json:"sortOrder" form:"sortOrder"`
}

// FaqStatusReq toggles a FAQ entry's active flag.
type FaqStatusReq struct {
	ID       uint `json:"id" form:"id" binding:"required"`
	IsActive bool `json:"isActive" form:"isActive"`
}

// TicketPageReq filters the ticket list.
type TicketPageReq struct {
	Page       int    `json:"page" form:"page"`
	PageSize   int    `json:"pageSize" form:"pageSize"`
	Status     string `json:"status" form:"status"`
	Priority   string `json:"priority" form:"priority"`
	Query      string `json:"query" form:"query"`
	CustomerID uint   `json:"customerId" form:"customerId"`
}

// MessagePageReq pages message history.
type MessagePageReq struct {
	Page     int   `json:"page" form:"page"`
	PageSize int   `json:"pageSize" form:"pageSize"`
	SinceSeq int64 `json:"sinceSeq" form:"sinceSeq"`
}

// PresenceQuery is the raw `ids` query string, e.g. "support:9,customer:51".
type PresenceQuery struct {
	Ids string `json:"ids" form:"ids" binding:"required"`
}

// ChatSettingsUpdateReq updates runtime support settings (DB wins over YAML).
type ChatSettingsUpdateReq struct {
	Values map[string]string `json:"values" form:"values" binding:"required"`
	Reason string            `json:"reason" form:"reason" binding:"max=500"`
}
