package support

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/model/common/response"
	supportModel "github.com/mo7amed4522/tayara_api/model/support"
	supportReq "github.com/mo7amed4522/tayara_api/model/support/request"
	supportResp "github.com/mo7amed4522/tayara_api/model/support/response"
	"gorm.io/gorm"
)

// FaqServices owns the admin-managed Q&A content and the append-only vote trail.
type FaqServices struct{}

// List returns active FAQ entries with tallies and the caller's current vote.
func (s FaqServices) List(customerID uint) ([]supportResp.FaqView, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	var faqs []supportModel.Faq
	if err := global.GVA_DB.Where("is_active = ?", true).Order("sort_order asc, id asc").Find(&faqs).Error; err != nil {
		return nil, err
	}
	out := make([]supportResp.FaqView, 0, len(faqs))
	for _, f := range faqs {
		out = append(out, s.view(f, customerID))
	}
	return out, nil
}

// view builds a FAQ view: tallies count only the latest vote per user.
func (s FaqServices) view(f supportModel.Faq, customerID uint) supportResp.FaqView {
	v := supportResp.FaqView{Faq: f, Tally: map[string]int{}}
	var votes []supportModel.FaqVote
	if err := global.GVA_DB.Where("faq_id = ?", f.ID).Order("id asc").Find(&votes).Error; err != nil {
		return v
	}
	latest := map[uint]string{}
	for _, vote := range votes {
		latest[vote.UserID] = vote.OptionKey
	}
	for uid, key := range latest {
		v.Tally[key]++
		if uid == customerID {
			v.MyVote = key
		}
	}
	return v
}

// Vote appends a vote. Choosing the same option again is a no-op; a different option appends a
// new row that supersedes the previous one (change-of-mind is allowed, mutation is not).
func (s FaqServices) Vote(customerID uint, faqID uint, req supportReq.VoteReq) error {
	if global.GVA_DB == nil {
		return errors.New("db not init")
	}
	var faq supportModel.Faq
	if err := global.GVA_DB.First(&faq, faqID).Error; err != nil {
		return response.NotFoundErr("faq not found")
	}
	if !faq.IsActive {
		return response.ConflictErr("faq is not active")
	}
	if !faqHasOption(faq, req.OptionKey) {
		return response.ValidationErr("invalid option")
	}
	var last supportModel.FaqVote
	err := global.GVA_DB.Where("faq_id = ? AND user_id = ?", faqID, customerID).Order("id desc").First(&last).Error
	if err == nil && last.OptionKey == req.OptionKey {
		return nil // idempotent
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if err := global.GVA_DB.Create(&supportModel.FaqVote{
		FaqID: faqID, UserID: customerID, OptionKey: req.OptionKey,
	}).Error; err != nil {
		return err
	}
	publishTicketEvent("faq.voted", 0, 0, uintString(faqID), map[string]any{
		"faq_id": faqID, "user_id": customerID, "option_key": req.OptionKey,
	})
	return nil
}

// faqHasOption reports whether the FAQ defines the option key. An FAQ with no options cannot be
// voted on. It round-trips through JSON so both []map[string]string (seeded) and
// []interface{} (decoded from a request) work.
func faqHasOption(faq supportModel.Faq, key string) bool {
	if faq.Options == nil {
		return false
	}
	raw, err := json.Marshal(faq.Options)
	if err != nil {
		return false
	}
	var parsed struct {
		Options []struct {
			Key string `json:"key"`
		} `json:"options"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return false
	}
	for _, o := range parsed.Options {
		if o.Key == key {
			return true
		}
	}
	return false
}

// ListAll returns every FAQ entry (admin).
func (s FaqServices) ListAll() ([]supportModel.Faq, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	var rows []supportModel.Faq
	err := global.GVA_DB.Order("sort_order asc, id asc").Find(&rows).Error
	return rows, err
}

// Upsert creates or updates a FAQ entry (admin).
func (s FaqServices) Upsert(req supportReq.FaqReq) (*supportModel.Faq, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	code := strings.TrimSpace(req.Code)
	if code == "" {
		return nil, response.ValidationErr("code is required")
	}
	var existing supportModel.Faq
	err := global.GVA_DB.Where("code = ?", code).First(&existing).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		row := supportModel.Faq{
			Code: code, Question: req.Question, QuestionAr: req.QuestionAr,
			Answer: req.Answer, AnswerAr: req.AnswerAr, Options: req.Options,
			IsActive: req.IsActive, SortOrder: req.SortOrder,
		}
		if err := global.GVA_DB.Create(&row).Error; err != nil {
			return nil, err
		}
		return &row, nil
	}
	existing.Question = req.Question
	existing.QuestionAr = req.QuestionAr
	existing.Answer = req.Answer
	existing.AnswerAr = req.AnswerAr
	existing.Options = req.Options
	existing.SortOrder = req.SortOrder
	if req.ID == 0 {
		existing.IsActive = req.IsActive
	}
	if err := global.GVA_DB.Save(&existing).Error; err != nil {
		return nil, err
	}
	return &existing, nil
}

// SetStatus toggles a FAQ entry's active flag (admin).
func (s FaqServices) SetStatus(req supportReq.FaqStatusReq) error {
	if global.GVA_DB == nil {
		return errors.New("db not init")
	}
	result := global.GVA_DB.Model(&supportModel.Faq{}).Where("id = ?", req.ID).Update("is_active", req.IsActive)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return response.NotFoundErr("faq not found")
	}
	return nil
}
