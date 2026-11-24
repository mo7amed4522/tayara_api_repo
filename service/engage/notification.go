package engage

import (
	"context"
	"errors"

	"github.com/mo7amed4522/tayara_api/constants"
	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/model/engage"
	engageReq "github.com/mo7amed4522/tayara_api/model/engage/request"
	"github.com/mo7amed4522/tayara_api/service/engage/dispatch"
	"github.com/mo7amed4522/tayara_api/utils"
	"go.uber.org/zap"
)

// NotificationServices records and delivers notifications. Templates drive the message body; the
// provider dispatchers (SMTP email / FCM push) do the actual delivery and mark sent/failed.
type NotificationServices struct{}

var errNotificationDeliveryDeferred = errors.New("notification delivery deferred")

// Emit records a notification for a recipient, then attempts provider delivery. If a template
// exists for (event_type, channel) it replaces the title/body. Delivery is best-effort: a send
// failure marks the record failed but never fails the calling flow.
func (n NotificationServices) Emit(eventType, recipientType string, recipientID uint, channel, title, body string) {
	n.emit(eventType, recipientType, recipientID, channel, title, body, "", true)
}

// EmitWithImage persists the notification, including its image URL, before attempting delivery.
func (n NotificationServices) EmitWithImage(eventType, recipientType string, recipientID uint, channel, title, body, imageURL string) {
	n.emit(eventType, recipientType, recipientID, channel, title, body, imageURL, false)
}

func (n NotificationServices) emit(eventType, recipientType string, recipientID uint, channel, title, body, imageURL string, useTemplate bool) {
	if global.GVA_DB == nil {
		global.GVA_LOG.Info("notify (db not init)", zap.String("event", eventType))
		return
	}
	// prefer the configured template
	if useTemplate {
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
		ImageURL:         imageURL,
	}
	if err := global.GVA_DB.Create(rec).Error; err != nil {
		global.GVA_LOG.Error("failed to emit notification", zap.Error(err))
		return
	}
	// deliver via the provider dispatcher
	if err := n.deliver(rec); err != nil {
		if errors.Is(err, errNotificationDeliveryDeferred) {
			return
		}
		_ = global.GVA_DB.Model(&engage.Notification{}).Where("id = ?", rec.ID).
			Update("status", constants.NotificationStatusFailed).Error
		return
	}
	_ = global.GVA_DB.Model(&engage.Notification{}).Where("id = ?", rec.ID).
		Update("status", constants.NotificationStatusSent).Error
}

// deliver routes a notification to the SMTP (email) or FCM (push) dispatcher.
func (n NotificationServices) deliver(rec *engage.Notification) error {
	ctx := context.Background()
	switch rec.Channel {
	case constants.NotifChannelEmail:
		var smtp dispatch.SmtpSender
		if !smtp.Enabled() {
			return nil // not configured -> treat as sent (no provider)
		}
		return smtp.Send([]string{recipientEmail(rec)}, rec.Title, rec.Body)
	case constants.NotifChannelPush:
		var fcm dispatch.FcmSender
		if !fcm.Enabled() {
			return errNotificationDeliveryDeferred
		}
		tokens, err := n.deviceTokens(rec.RecipientID)
		if err != nil {
			return err
		}
		if len(tokens) == 0 {
			return errNotificationDeliveryDeferred
		}
		return fcm.Send(ctx, tokens, rec.Title, rec.Body, rec.ImageURL)
	default:
		return nil
	}
}

// deviceTokens returns the recipient's active FCM tokens (push delivery targets).
func (n NotificationServices) deviceTokens(userID uint) ([]string, error) {
	var rows []engage.DeviceToken
	if err := global.GVA_DB.Where("user_id = ? AND is_active = ?", userID, true).Find(&rows).Error; err != nil {
		return nil, err
	}

	var pushTokens []engage.PushToken
	if err := global.GVA_DB.Where("user_id = ? AND is_active = ?", userID, true).Find(&pushTokens).Error; err != nil {
		return nil, err
	}

	out := make([]string, 0, len(rows)+len(pushTokens))
	seen := make(map[string]struct{}, len(rows)+len(pushTokens))
	for _, r := range rows {
		if _, exists := seen[r.Token]; exists {
			continue
		}
		seen[r.Token] = struct{}{}
		out = append(out, r.Token)
	}
	for _, r := range pushTokens {
		if _, exists := seen[r.Token]; exists {
			continue
		}
		seen[r.Token] = struct{}{}
		out = append(out, r.Token)
	}
	return out, nil
}

// recipientEmail resolves a recipient's email address. For email delivery the recipient id is a
// user id; we resolve it from the bx user profile. If none found, delivery is skipped.
func recipientEmail(rec *engage.Notification) string {
	var profile struct {
		Email string
	}
	_ = global.GVA_DB.Table("bx_user_profile").Select("email").Where("id = ?", rec.RecipientID).Scan(&profile).Error
	return profile.Email
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
func (n NotificationServices) ListNotifications(req engageReq.NotificationPageReq, recipientType string, recipientID uint) ([]engage.Notification, int64, int64, error) {
	if global.GVA_DB == nil {
		return nil, 0, 0, errors.New("db not init")
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
	var list []engage.Notification
	pr, err := utils.PaginateWithCustomQuery(db, db.Order("created_at desc"), req.Page, req.PageSize, &list)
	if err != nil {
		return nil, 0, 0, err
	}
	return list, pr.Total, pr.TotalPage, nil
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

// ListTemplates returns a paginated list of notification templates.
func (n NotificationServices) ListTemplates(page, pageSize int) ([]engage.NotificationTemplate, int64, int64, error) {
	if global.GVA_DB == nil {
		return nil, 0, 0, errors.New("db not init")
	}
	var list []engage.NotificationTemplate
	countDB := global.GVA_DB.Model(&engage.NotificationTemplate{})
	queryDB := global.GVA_DB.Model(&engage.NotificationTemplate{}).
		Order("event_type asc, channel asc")
	pr, err := utils.PaginateWithCustomQuery(countDB, queryDB, page, pageSize, &list)
	if err != nil {
		return nil, 0, 0, err
	}
	return list, pr.Total, pr.TotalPage, nil
}
