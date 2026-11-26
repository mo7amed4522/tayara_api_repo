package support

import (
	"errors"
	"time"

	"github.com/mo7amed4522/tayara_api/global"
	supportModel "github.com/mo7amed4522/tayara_api/model/support"
)

// OutboxServices re-publishes messages whose publish attempt failed. The message table is the
// outbox: published_at NULL marks work to do, and consumers dedupe on message_uuid.
type OutboxServices struct{}

// Sweep publishes up to 200 unpublished messages, oldest first.
func (s OutboxServices) Sweep() error {
	if global.GVA_DB == nil {
		return errors.New("db not init")
	}
	var rows []supportModel.Message
	if err := global.GVA_DB.Where("published_at IS NULL").Order("id asc").Limit(200).Find(&rows).Error; err != nil {
		return err
	}
	for _, m := range rows {
		view, err := (MessageServices{}).viewFor(m)
		if err != nil {
			return err
		}
		if !publishMessage(*view, m.ConversationID, m.TicketID) {
			continue
		}
		_ = global.GVA_DB.Model(&supportModel.Message{}).Where("id = ?", m.ID).
			UpdateColumn("published_at", time.Now().UTC()).Error
	}
	return nil
}
