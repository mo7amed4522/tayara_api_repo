package engage

import (
	"github.com/mo7amed4522/tayara_api/constants"
	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/model/engage"
	"github.com/mo7amed4522/tayara_api/utils"
	"go.uber.org/zap"
)

type NotificationServices struct{}

// Emit records a notification for a recipient. Once the push/email providers are wired (Phase 6)
// this will also dispatch delivery; today it persists the pending log row (and logs on failure).
func (n NotificationServices) Emit(eventType, recipientType string, recipientID uint, channel, title, body string) {
	if global.GVA_DB == nil {
		global.GVA_LOG.Info("notify (db not init)", zap.String("event", eventType), zap.String("title", title))
		return
	}
	rec := &engage.Notification{
		NotificationCode: utils.NextCode(constants.CodePrefixNotification),
		EventType:        eventType,
		RecipientType:    recipientType,
		RecipientID:      recipientID,
		Channel:          channel,
		Status:           constants.NotificationStatusPending,
		Title:            title,
		Body:             body,
	}
	if err := global.GVA_DB.Create(rec).Error; err != nil {
		global.GVA_LOG.Error("failed to emit notification", zap.Error(err))
	}
}
