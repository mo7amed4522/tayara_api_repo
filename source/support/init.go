// Package support seeds the runtime chat settings, demo FAQ content and the MySQL append-only
// triggers for the support module. Tables are created by AutoMigrate (initialize/gorm_biz.go);
// this initializer only seeds rows and (on MySQL) installs DB-level immutability guards.
package support

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/mo7amed4522/tayara_api/constants"
	"github.com/mo7amed4522/tayara_api/global"
	basemodel "github.com/mo7amed4522/tayara_api/model/base"
	"github.com/mo7amed4522/tayara_api/model/common"
	supportmodel "github.com/mo7amed4522/tayara_api/model/support"
	"github.com/mo7amed4522/tayara_api/service/system"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// initOrderSupport runs after the demo domain seeds so the settings/FAQ rows land last.
const initOrderSupport = system.InitOrderExternal + 9

type supportInit struct{}

func init() {
	system.RegisterInit(initOrderSupport, &supportInit{})
}

func (s *supportInit) InitializerName() string { return "support_chat_settings_and_faq" }

// MigrateTable/TableCreated are no-ops: tables migrate via gorm_biz AutoMigrate.
func (s *supportInit) MigrateTable(ctx context.Context) (context.Context, error) { return ctx, nil }
func (s *supportInit) TableCreated(ctx context.Context) bool                     { return true }

func (s *supportInit) DataInserted(ctx context.Context) bool {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok || db == nil {
		return false
	}
	var n int64
	if err := db.Model(&basemodel.SystemSetting{}).
		Where("key = ?", constants.SupportSettingChatEnabled).Count(&n).Error; err != nil {
		return false
	}
	return n > 0
}

func (s *supportInit) InitializeData(ctx context.Context) (context.Context, error) {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok || db == nil {
		return ctx, system.ErrMissingDBContext
	}
	if err := seedSettings(db); err != nil {
		return ctx, err
	}
	if err := seedFaqs(db); err != nil {
		return ctx, err
	}
	installTriggers(db)
	return ctx, nil
}

// seedSettings writes the deploy-time chat policy into base_system_setting so an admin can edit
// it at runtime (DB value wins over YAML). Existing rows are never overwritten.
func seedSettings(db *gorm.DB) error {
	cfg := global.GVA_CONFIG.Support
	rows := []basemodel.SystemSetting{
		{Section: constants.SupportSettingSection, Key: constants.SupportSettingChatEnabled,
			Value: strconv.FormatBool(cfg.Chat.Enabled), Description: "Support chat enabled"},
		{Section: constants.SupportSettingSection, Key: constants.SupportSettingWelcomeMessage,
			Value: cfg.Chat.WelcomeMessage, Description: "Welcome message (en)"},
		{Section: constants.SupportSettingSection, Key: constants.SupportSettingWelcomeMessageAr,
			Value: cfg.Chat.WelcomeMessageAr, Description: "Welcome message (ar)"},
		{Section: constants.SupportSettingSection, Key: constants.SupportSettingAllowedSenderTypes,
			Value: strings.Join(cfg.Chat.AllowedSenderTypes, ","), Description: "Allowed sender types (CSV)"},
		{Section: constants.SupportSettingSection, Key: constants.SupportSettingMaxMessageBytes,
			Value: strconv.Itoa(cfg.Chat.MaxMessageBytes), Description: "Max message bytes"},
		{Section: constants.SupportSettingSection, Key: constants.SupportSettingRateLimitPer10s,
			Value: strconv.Itoa(cfg.Chat.RateLimit.Per10s), Description: "Messages per 10 seconds"},
		{Section: constants.SupportSettingSection, Key: constants.SupportSettingRateLimitPerHour,
			Value: strconv.Itoa(cfg.Chat.RateLimit.PerHour), Description: "Messages per hour"},
		{Section: constants.SupportSettingSection, Key: constants.SupportSettingEscalateAfterMinutes,
			Value: strconv.Itoa(cfg.Chat.Assignment.EscalateAfterMinutesUnanswered), Description: "Escalate after unanswered minutes"},
		{Section: constants.SupportSettingSection, Key: constants.SupportSettingRequireFaqBeforeChat,
			Value: strconv.FormatBool(cfg.Chat.RequireFaqBeforeChat), Description: "Require FAQ before chat"},
		{Section: constants.SupportSettingSection, Key: constants.SupportSettingAttachmentsEnabled,
			Value: strconv.FormatBool(cfg.Chat.Attachments.Enabled), Description: "Attachments enabled"},
		{Section: constants.SupportSettingSection, Key: constants.SupportSettingAttachmentsMaxFileBytes,
			Value: strconv.FormatInt(cfg.Chat.Attachments.MaxFileBytes, 10), Description: "Max bytes per attachment"},
		{Section: constants.SupportSettingSection, Key: constants.SupportSettingAttachmentsMaxFiles,
			Value: strconv.Itoa(cfg.Chat.Attachments.MaxFilesPerMessage), Description: "Max attachments per message"},
		{Section: constants.SupportSettingSection, Key: constants.SupportSettingAttachmentsMaxTotal,
			Value: strconv.FormatInt(cfg.Chat.Attachments.MaxTotalBytesPerMessage, 10), Description: "Max attachment bytes per message"},
		{Section: constants.SupportSettingSection, Key: constants.SupportSettingAttachmentsAllowedMime,
			Value: strings.Join(cfg.Chat.Attachments.AllowedMime, ","), Description: "Allowed attachment MIME types (CSV)"},
		{Section: constants.SupportSettingSection, Key: constants.SupportSettingLastSeenDisplay,
			Value: strconv.FormatBool(cfg.Presence.LastSeenDisplay), Description: "Show last seen"},
		{Section: constants.SupportSettingSection, Key: constants.SupportSettingHideLastSeenForRoles,
			Value: strings.Join(cfg.Presence.HideLastSeenForRoles, ","), Description: "Roles whose last seen is hidden (CSV)"},
	}
	for i := range rows {
		rows[i].IsActive = true
	}
	return db.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "key"}}, DoNothing: true}).
		Create(&rows).Error
}

func seedFaqs(db *gorm.DB) error {
	faqs := []supportmodel.Faq{
		{
			Code:       "how-to-cancel-booking",
			Question:   "How do I cancel a booking?",
			QuestionAr: "كيف ألغي الحجز؟",
			Answer:     "Open the booking, choose Cancel, and confirm the reason. Cancellation terms depend on the package policy.",
			AnswerAr:   "افتح الحجز ثم اختر إلغاء وأكد السبب. تعتمد شروط الإلغاء على سياسة الباقة.",
			Options:    common.JSONMap{},
			IsActive:   true,
			SortOrder:  1,
		},
		{
			Code:       "refund-timeline",
			Question:   "When will I receive my refund?",
			QuestionAr: "متى أستلم المبلغ المسترد؟",
			Answer:     "Refunds are issued to the original payment method and usually arrive within 5–10 business days.",
			AnswerAr:   "يتم رد المبلغ إلى وسيلة الدفع الأصلية ويصل عادة خلال 5–10 أيام عمل.",
			Options: common.JSONMap{"options": []map[string]string{
				{"key": constants.SupportVoteHelpful, "label": "Helpful"},
				{"key": constants.SupportVoteNotHelpful, "label": "Not helpful"},
			}},
			IsActive:  true,
			SortOrder: 2,
		},
		{
			Code:       "contact-support",
			Question:   "How do I reach a support specialist?",
			QuestionAr: "كيف أتواصل مع مختص الدعم؟",
			Answer:     "Open a support ticket from this page; a specialist replies in the same chat thread.",
			AnswerAr:   "افتح تذكرة دعم من هذه الصفحة وسيرد المختص في نفس المحادثة.",
			Options:    common.JSONMap{},
			IsActive:   true,
			SortOrder:  3,
		},
	}
	return db.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "code"}}, DoNothing: true}).Create(&faqs).Error
}

// installTriggers adds MySQL BEFORE UPDATE/DELETE guards. On other dialects (SQLite in tests)
// it is a no-op. Failures are logged, never fatal: the GORM hooks remain the primary guard.
func installTriggers(db *gorm.DB) {
	if db.Dialector == nil || db.Dialector.Name() != "mysql" {
		return
	}
	stmts := []string{}
	appendOnly := []struct{ table, msg string }{
		{"support_message", "support_message is append-only"},
		{"support_faq_vote", "support_faq_vote is append-only"},
		{"support_assignment_history", "support_assignment_history is append-only"},
		{"support_review", "support_review is append-only"},
		{"support_presence", "support_presence is append-only"},
	}
	for _, t := range appendOnly {
		stmts = append(stmts,
			fmt.Sprintf("DROP TRIGGER IF EXISTS trg_%s_no_update", t.table),
			fmt.Sprintf("CREATE TRIGGER trg_%s_no_update BEFORE UPDATE ON %s FOR EACH ROW SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = '%s'", t.table, t.table, t.msg),
			fmt.Sprintf("DROP TRIGGER IF EXISTS trg_%s_no_delete", t.table),
			fmt.Sprintf("CREATE TRIGGER trg_%s_no_delete BEFORE DELETE ON %s FOR EACH ROW SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = '%s'", t.table, t.table, t.msg),
		)
	}
	stmts = append(stmts,
		"DROP TRIGGER IF EXISTS trg_support_attachment_update",
		`CREATE TRIGGER trg_support_attachment_update BEFORE UPDATE ON support_attachment
FOR EACH ROW
BEGIN
    IF NOT (
        (OLD.message_id IS NULL AND NEW.message_id IS NOT NULL
         AND OLD.status = 'uploaded' AND NEW.status = 'attached')
        OR (OLD.message_id IS NULL AND NEW.message_id IS NULL
            AND OLD.status = 'uploaded' AND NEW.status = 'rejected')
    )
    OR NOT (OLD.attachment_uuid <=> NEW.attachment_uuid AND OLD.conversation_id <=> NEW.conversation_id
        AND OLD.ticket_id <=> NEW.ticket_id AND OLD.uploader_type <=> NEW.uploader_type
        AND OLD.uploader_id <=> NEW.uploader_id AND OLD.kind <=> NEW.kind
        AND OLD.mime_type <=> NEW.mime_type AND OLD.size_bytes <=> NEW.size_bytes
        AND OLD.original_name <=> NEW.original_name AND OLD.storage_key <=> NEW.storage_key
        AND OLD.storage_name <=> NEW.storage_name AND OLD.sha256 <=> NEW.sha256
        AND OLD.width <=> NEW.width AND OLD.height <=> NEW.height
        AND OLD.thumbnail_key <=> NEW.thumbnail_key AND OLD.created_at <=> NEW.created_at)
    THEN SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'support_attachment only permits the one-time link';
    END IF;
END`,
		"DROP TRIGGER IF EXISTS trg_support_attachment_no_delete",
		"CREATE TRIGGER trg_support_attachment_no_delete BEFORE DELETE ON support_attachment FOR EACH ROW SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'support_attachment is append-only'",
	)
	for _, stmt := range stmts {
		if err := db.Exec(stmt).Error; err != nil && global.GVA_LOG != nil {
			global.GVA_LOG.Warn("support trigger install failed", zap.Error(err))
		}
	}
}
