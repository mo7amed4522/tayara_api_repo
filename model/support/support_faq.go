package support

import (
	"errors"
	"time"

	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/model/common"
	"gorm.io/gorm"
)

// Faq admin-managed Q&A content. `options` may hold a poll definition; an empty map means the
// entry is informational only. This table is editable content, not append-only.
type Faq struct {
	ID         uint           `json:"id" gorm:"primarykey;comment:ID"`
	Code       string         `json:"code" form:"code" gorm:"size:64;uniqueIndex;comment:Stable FAQ code"`
	Question   string         `json:"question" form:"question" gorm:"size:500;comment:Question (en)"`
	QuestionAr string         `json:"questionAr" form:"questionAr" gorm:"size:500;comment:Question (ar)"`
	Answer     string         `json:"answer" form:"answer" gorm:"type:text;comment:Answer (en)"`
	AnswerAr   string         `json:"answerAr" form:"answerAr" gorm:"type:text;comment:Answer (ar)"`
	Options    common.JSONMap `json:"options" form:"options" gorm:"comment:Poll options; empty = informational"`
	IsActive   bool           `json:"isActive" form:"isActive" gorm:"default:true;index;comment:Is active"`
	SortOrder  int            `json:"sortOrder" form:"sortOrder" gorm:"default:0;comment:Display order"`
	global.GVA_MODEL
}

func (Faq) TableName() string { return "support_faq" }

// FaqVote an append-only vote record. Changing your mind appends a new row; the current vote is
// the latest row per (faq_id, user_id) and the tally counts only those.
type FaqVote struct {
	ID        uint      `json:"id" gorm:"primarykey;comment:ID"`
	FaqID     uint      `json:"faqId" form:"faqId" gorm:"index:idx_faq_user_id;comment:FAQ ID"`
	UserID    uint      `json:"userId" form:"userId" gorm:"index:idx_faq_user_id;comment:Voting customer ID"`
	OptionKey string    `json:"optionKey" form:"optionKey" gorm:"size:64;comment:Chosen option key"`
	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime;<-:create;comment:Server created at"`
}

func (FaqVote) TableName() string { return "support_faq_vote" }

func (FaqVote) BeforeUpdate(tx *gorm.DB) error { return errors.New("support_faq_vote is append-only") }
func (FaqVote) BeforeDelete(tx *gorm.DB) error { return errors.New("support_faq_vote is append-only") }
