package engage

import "github.com/mo7amed4522/tayara_api/global"

// Notification an emitted notification log row (event type, recipient, channel, delivery status).
// Phase 2 provides the sink; the fuller engage module (templates, resend, ads) lands in Phase 6.
type Notification struct {
	ID               uint   `json:"id" gorm:"primarykey;comment:ID"`
	NotificationCode string `json:"notificationCode" form:"notificationCode" gorm:"size:64;uniqueIndex;comment:NT- prefixed code"`
	EventType        string `json:"eventType" form:"eventType" gorm:"size:64;index;comment:Event type"`
	RecipientType    string `json:"recipientType" form:"recipientType" gorm:"size:16;index;comment:user|agent|admin"`
	RecipientID      uint   `json:"recipientId" form:"recipientId" gorm:"index;comment:Recipient ID"`
	Channel          string `json:"channel" form:"channel" gorm:"size:16;comment:push|email"`
	Status           string `json:"status" form:"status" gorm:"size:16;default:pending;comment:pending|sent|failed"`
	Title            string `json:"title" form:"title" gorm:"size:191;comment:Title"`
	Body             string `json:"body" form:"body" gorm:"size:500;comment:Body"`
	global.GVA_MODEL
}

func (Notification) TableName() string { return "engage_notification" }
