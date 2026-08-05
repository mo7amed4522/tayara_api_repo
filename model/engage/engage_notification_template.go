package engage

import "github.com/mo7amed4522/tayara_api/global"

// NotificationTemplate maps an event type + channel to a message, so notifications are
// configurable without code changes. The Emit path looks up a template by event_type.
type NotificationTemplate struct {
	ID        uint   `json:"id" gorm:"primarykey;comment:ID"`
	EventType string `json:"eventType" form:"eventType" gorm:"size:64;index;comment:Event type"`
	Channel   string `json:"channel" form:"channel" gorm:"size:16;index;comment:push|email"`
	Subject   string `json:"subject" form:"subject" gorm:"size:191;comment:Subject/title"`
	Body      string `json:"body" form:"body" gorm:"size:500;comment:Message body"`
	IsActive  bool   `json:"isActive" form:"isActive" gorm:"default:true;comment:Is active"`
	global.GVA_MODEL
}

func (NotificationTemplate) TableName() string { return "engage_notification_template" }
