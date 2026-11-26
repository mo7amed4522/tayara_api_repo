package support

import (
	"time"

	"github.com/mo7amed4522/tayara_api/global"
)

// Ticket a support ticket. Optionally anchors to a booking or a dispute, but a free-standing
// ticket (both nil) is valid. This is the mutable aggregate; its message history is append-only.
type Ticket struct {
	ID                  uint       `json:"id" gorm:"primarykey;comment:ID"`
	TicketCode          string     `json:"ticketCode" form:"ticketCode" gorm:"size:64;uniqueIndex;comment:ST- prefixed code"`
	Subject             string     `json:"subject" form:"subject" gorm:"size:255;comment:Ticket subject"`
	CustomerID          uint       `json:"customerId" form:"customerId" gorm:"index;comment:Customer (bx_user_profile) ID"`
	AssignedEmployeeID  *uint      `json:"assignedEmployeeId" form:"assignedEmployeeId" gorm:"index;comment:Assigned sys_admin_profile ID"`
	AssignedAuthorityID *uint      `json:"assignedAuthorityId" form:"assignedAuthorityId" gorm:"comment:4=support 5=quality"`
	BookingID           *uint      `json:"bookingId" form:"bookingId" gorm:"index;comment:Optional bookings.id link"`
	DisputeID           *uint      `json:"disputeId" form:"disputeId" gorm:"index;comment:Optional disputes.id link"`
	Status              string     `json:"status" form:"status" gorm:"size:24;index;default:open;comment:open|pending_customer|pending_support|solved|not_solved|closed"`
	Priority            string     `json:"priority" form:"priority" gorm:"size:16;default:normal;comment:low|normal|high|urgent"`
	LastMessageAt       *time.Time `json:"lastMessageAt" form:"lastMessageAt" gorm:"index;comment:Last message at (inbox sorting)"`
	ClosedAt            *time.Time `json:"closedAt" form:"closedAt" gorm:"comment:Closed at"`
	global.GVA_MODEL
}

func (Ticket) TableName() string { return "support_ticket" }

// Conversation one conversation per ticket. `last_seq` is the last allocated sequence number and
// `last_body_hash` is the tail of the tamper-evident hash chain.
type Conversation struct {
	ID                 uint       `json:"id" gorm:"primarykey;comment:ID"`
	TicketID           uint       `json:"ticketId" form:"ticketId" gorm:"uniqueIndex;comment:Ticket ID (one conversation per ticket)"`
	Status             string     `json:"status" form:"status" gorm:"size:16;index;default:active;comment:active|closed"`
	OpenedAt           *time.Time `json:"openedAt" form:"openedAt" gorm:"comment:Opened at"`
	ClosedAt           *time.Time `json:"closedAt" form:"closedAt" gorm:"comment:Closed at"`
	FirstMessageID     *uint      `json:"firstMessageId" form:"firstMessageId" gorm:"comment:Welcome message ID"`
	MessageCount       int        `json:"messageCount" form:"messageCount" gorm:"default:0;comment:Monotonic message counter"`
	LastSeq            int64      `json:"lastSeq" form:"lastSeq" gorm:"default:0;comment:Last allocated seq"`
	LastBodyHash       string     `json:"lastBodyHash" form:"lastBodyHash" gorm:"size:64;comment:Hash chain tail"`
	AssignedEmployeeID *uint      `json:"assignedEmployeeId" form:"assignedEmployeeId" gorm:"index;comment:Mirrors ticket assignee"`
	global.GVA_MODEL
}

func (Conversation) TableName() string { return "support_conversation" }
