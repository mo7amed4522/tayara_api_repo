package engage

import (
	"errors"

	"github.com/mo7amed4522/tayara_api/constants"
	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/model/engage"
	engageReq "github.com/mo7amed4522/tayara_api/model/engage/request"
	"github.com/mo7amed4522/tayara_api/utils"
	"go.uber.org/zap"
)

// NotificationServices records and resend notifications. Phase 6 adds template-driven messages
// and the resend path; provider delivery (push/email) dispatchers are wired once available.
type NotificationServices struct{}

// Emit records a notification for a recipient. If a template exists for (event_type, channel)
// it replaces the title/body; otherwise the provided title/body are used verbatim.
func (n NotificationServices) Emit(eventType, recipientType string, recipientID uint, channel, title, body string) {
	if global.GVA_DB == nil {
		global.GVA_LOG.Info("notify (db not init)", zap.String("event", eventType))
		return
	}
	// prefer the configured template
	var tpl engage.NotificationTemplate
	if err := global.GVA_DB.Where("event_type = ? AND channel = ? AND is_active = ?", eventType, channel, true).
		First(&tpl).Error; err == nil {
		if tpl.Subject != "" {
			title = tpl.Subject
		}
		if tpl.Body != "" {
			body = tpl.Body
		}
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

// Resend re-emits an existing notification by id (marks it pending/sent again).
func (n NotificationServices) Resend(id uint) (*engage.Notification, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	var rec engage.Notification
	if err := global.GVA_DB.First(&rec, id).Error; err != nil {
		return nil, errors.New("notification not found")
	}
	if err := global.GVA_DB.Model(&engage.Notification{}).Where("id = ?", id).
		Update("status", constants.NotificationStatusPending).Error; err != nil {
		return nil, err
	}
	rec.Status = constants.NotificationStatusPending
	return &rec, nil
}

// ListNotifications returns notifications filtered by recipient/status (admin sees all).
func (n NotificationServices) ListNotifications(req engageReq.NotificationPageReq, recipientType string, recipientID uint) ([]engage.Notification, int64, error) {
	if global.GVA_DB == nil {
		return nil, 0, errors.New("db not init")
	}
	db := global.GVA_DB.Model(&engage.Notification{})
	if recipientType != "" {
		db = db.Where("recipient_type = ?", recipientType)
	}
	if recipientID > 0 {
		db = db.Where("recipient_id = ?", recipientID)
	}
	if req.Status != "" {
		db = db.Where("status = ?", req.Status)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []engage.Notification
	if err := db.Scopes(req.Paginate()).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// --- notification templates ---

// UpsertTemplate creates or updates a notification template (keyed by event_type + channel).
func (n NotificationServices) UpsertTemplate(req engageReq.NotificationTemplateReq) (*engage.NotificationTemplate, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	if req.ID > 0 {
		var exists engage.NotificationTemplate
		if err := global.GVA_DB.First(&exists, req.ID).Error; err != nil {
			return nil, errors.New("template not found")
		}
		err := global.GVA_DB.Model(&engage.NotificationTemplate{}).Where("id = ?", req.ID).Updates(map[string]interface{}{
			"event_type": req.EventType, "channel": req.Channel, "subject": req.Subject,
			"body": req.Body, "is_active": req.IsActive,
		}).Error
		if err != nil {
			return nil, err
		}
		exists.EventType, exists.Channel, exists.Subject, exists.Body, exists.IsActive =
			req.EventType, req.Channel, req.Subject, req.Body, req.IsActive
		return &exists, nil
	}
	tpl := &engage.NotificationTemplate{
		EventType: req.EventType, Channel: req.Channel, Subject: req.Subject, Body: req.Body, IsActive: req.IsActive,
	}
	if err := global.GVA_DB.Create(tpl).Error; err != nil {
		return nil, err
	}
	return tpl, nil
}

// DeleteTemplate removes a notification template.
func (n NotificationServices) DeleteTemplate(id uint) error {
	if global.GVA_DB == nil {
		return errors.New("db not init")
	}
	return global.GVA_DB.Delete(&engage.NotificationTemplate{}, id).Error
}
