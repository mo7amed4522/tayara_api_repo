package recommendation

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/mo7amed4522/tayara_api/constants"
	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/model/catalog"
	"github.com/mo7amed4522/tayara_api/model/dxapi"
	"github.com/mo7amed4522/tayara_api/model/recommendation"
	"github.com/mo7amed4522/tayara_api/model/recommendation/request"
	recResp "github.com/mo7amed4522/tayara_api/model/recommendation/response"
	"github.com/mo7amed4522/tayara_api/utils"
	"github.com/mo7amed4522/tayara_api/utils/cache"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// RecommendationServices owns stored recommendations: admin CRUD/review, the public approved list,
// settings, engagement events and the generation engine (see recommendation_engine.go).
type RecommendationServices struct{}

// RecommendationServiceApp the shared service instance for APIs and the scheduled job.
var RecommendationServiceApp = new(RecommendationServices)

const (
	// publicCacheTTL how long the approved public list is cached.
	publicCacheTTL = 5 * time.Minute
	// defaultMaxItems used when a setting does not configure a limit.
	defaultMaxItems = 10
	// maxReasonLength the stored reason is truncated to this many runes.
	maxReasonLength = 255
	// publicPoolLimit how many approved rows are cached and paged for the public list.
	publicPoolLimit = 200
	// maxPublicPageSize the largest page the public list serves.
	maxPublicPageSize = 50
)

// DefaultPromptTemplate the versioned system prompt seeded for every placement.
const DefaultPromptTemplate = `You are the recommendation engine of Tayara, a travel marketplace.
You receive a JSON list of CANDIDATES (packages, trips, agencies) and a PLACEMENT.
Pick the best items for that placement using rating, popularity, recency, price fit, and the click/booking signals.
Rules:
- Use ONLY ids that appear in CANDIDATES. Never invent items.
- Return at most {{max_items}} items, best first.
- Mix target types if the placement allows it. Do not return the same agency more than twice.
- "reason" is one short, factual sentence for the admin and customer. No marketing exaggeration.
- Output JSON only, matching the schema. No prose, no markdown.`

// ---------------------------------------------------------------------------
// Admin: list / find
// ---------------------------------------------------------------------------

// @author:		[mo7amed4522](https://github.com/mo7amed4522)
// @function:		GetRecommendationList
// @description:	Paginated admin list of stored recommendations with filters
// @param:			info request.RecommendationSearch
// @return:		list []recommendation.Recommendation, total int64, err error
func (s *RecommendationServices) GetRecommendationList(info request.RecommendationSearch) (list []recommendation.Recommendation, total int64, err error) {
	if global.GVA_DB == nil {
		return nil, 0, errors.New("db not init")
	}
	db := global.GVA_DB.Model(&recommendation.Recommendation{})
	if info.TargetType != "" {
		db = db.Where("target_type = ?", info.TargetType)
	}
	if info.Placement != "" {
		db = db.Where("placement = ?", info.Placement)
	}
	if info.Audience != "" {
		db = db.Where("audience = ?", info.Audience)
	}
	if info.Status != "" {
		db = db.Where("status = ?", info.Status)
	}
	if info.Source != "" {
		db = db.Where("source = ?", info.Source)
	}
	if info.Keyword != "" {
		db = db.Where("reason LIKE ?", "%"+info.Keyword+"%")
	}
	if from, ok := parseQueryTime(info.StartTime, false); ok {
		db = db.Where("created_at >= ?", from)
	}
	if to, ok := parseQueryTime(info.EndTime, true); ok {
		db = db.Where("created_at <= ?", to)
	}

	var rows []recommendation.Recommendation
	pr, err := utils.PaginateWithCustomQuery(
		db.Session(&gorm.Session{}),
		db.Session(&gorm.Session{}).Order("is_pinned desc, priority desc, score desc"),
		info.Page, info.PageSize, &rows,
	)
	if err != nil {
		return nil, 0, err
	}
	return rows, pr.Total, nil
}

// @author:		[mo7amed4522](https://github.com/mo7amed4522)
// @function:		FindRecommendation
// @description:	Get one stored recommendation by id
// @param:			id uint
// @return:		rec *recommendation.Recommendation, err error
func (s *RecommendationServices) FindRecommendation(id uint) (rec *recommendation.Recommendation, err error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	var row recommendation.Recommendation
	if err = global.GVA_DB.First(&row, id).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

// ---------------------------------------------------------------------------
// Admin: create / update / review / delete
// ---------------------------------------------------------------------------

// @author:		[mo7amed4522](https://github.com/mo7amed4522)
// @function:		CreateRecommendation
// @description:	Manual admin entry; the target must exist and be active
// @param:			req request.CreateRecommendationReq, adminID uint
// @return:		rec *recommendation.Recommendation, err error
func (s *RecommendationServices) CreateRecommendation(req request.CreateRecommendationReq, adminID uint) (rec *recommendation.Recommendation, err error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	if err := s.ensureTarget(req.TargetType, req.TargetID); err != nil {
		return nil, err
	}
	audience := req.Audience
	if audience == "" {
		audience = constants.RecommendationDefaultAudience
	}
	score := clampScore(req.Score)
	if req.Score == 0 {
		score = 1
	}
	now := time.Now()
	row := recommendation.Recommendation{
		TargetType: req.TargetType, TargetID: req.TargetID, Placement: req.Placement, Audience: audience,
		Source: constants.RecommendationSourceAdmin, Status: constants.RecommendationStatusApproved, Score: score,
		Reason: truncateReason(req.Reason), ReasonAr: truncateReason(req.ReasonAr),
		Priority: req.Priority, IsPinned: req.IsPinned, StartsAt: req.StartsAt, EndsAt: req.EndsAt,
		ReviewedBy: adminID, ReviewedAt: &now,
	}
	var existing recommendation.Recommendation
	if err := global.GVA_DB.Unscoped().
		Where("target_type = ? AND target_id = ? AND placement = ? AND audience = ?", row.TargetType, row.TargetID, row.Placement, row.Audience).
		First(&existing).Error; err == nil {
		if existing.DeletedAt.Valid {
			row.ID = existing.ID
			if err := global.GVA_DB.Unscoped().Model(&existing).Updates(map[string]interface{}{
				"deleted_at": nil, "source": row.Source, "status": row.Status, "score": row.Score,
				"reason": row.Reason, "reason_ar": row.ReasonAr, "priority": row.Priority,
				"is_pinned": row.IsPinned, "starts_at": row.StartsAt, "ends_at": row.EndsAt,
				"reviewed_by": row.ReviewedBy, "reviewed_at": row.ReviewedAt,
			}).Error; err != nil {
				return nil, err
			}
			row = existing
			return &row, nil
		}
		return nil, fmt.Errorf("a recommendation for %s %d already exists in placement %q", row.TargetType, row.TargetID, row.Placement)
	}
	if err := global.GVA_DB.Create(&row).Error; err != nil {
		return nil, err
	}
	s.invalidatePlacementCache(row.Placement, row.Audience)
	return &row, nil
}

// @author:		[mo7amed4522](https://github.com/mo7amed4522)
// @function:		UpdateRecommendation
// @description:	Admin edit of the allowed fields only
// @param:			req request.UpdateRecommendationReq
// @return:		err error
func (s *RecommendationServices) UpdateRecommendation(req request.UpdateRecommendationReq) (err error) {
	if global.GVA_DB == nil {
		return errors.New("db not init")
	}
	var row recommendation.Recommendation
	if err := global.GVA_DB.First(&row, req.ID).Error; err != nil {
		return err
	}
	updates := map[string]interface{}{}
	if req.Reason != nil {
		updates["reason"] = truncateReason(*req.Reason)
	}
	if req.ReasonAr != nil {
		updates["reason_ar"] = truncateReason(*req.ReasonAr)
	}
	if req.Priority != nil {
		updates["priority"] = *req.Priority
	}
	if req.IsPinned != nil {
		updates["is_pinned"] = *req.IsPinned
	}
	if req.StartsAt != nil {
		updates["starts_at"] = req.StartsAt
	}
	if req.EndsAt != nil {
		updates["ends_at"] = req.EndsAt
	}
	if req.Placement != "" {
		updates["placement"] = req.Placement
	}
	if req.Audience != "" {
		updates["audience"] = req.Audience
	}
	if len(updates) == 0 {
		return errors.New("no fields to update")
	}
	if err := global.GVA_DB.Model(&row).Updates(updates).Error; err != nil {
		return err
	}
	s.invalidatePlacementCache(row.Placement, row.Audience)
	if req.Placement != "" {
		s.invalidatePlacementCache(req.Placement, req.Audience)
	}
	return nil
}

// @author:		[mo7amed4522](https://github.com/mo7amed4522)
// @function:		ReviewRecommendation
// @description:	Approve or reject one recommendation as an admin
// @param:			id uint, reviewerID uint, action string
// @return:		err error
func (s *RecommendationServices) ReviewRecommendation(id, reviewerID uint, action string) (err error) {
	status, err := reviewStatus(action)
	if err != nil {
		return err
	}
	if global.GVA_DB == nil {
		return errors.New("db not init")
	}
	var row recommendation.Recommendation
	if err := global.GVA_DB.First(&row, id).Error; err != nil {
		return err
	}
	now := time.Now()
	if err := global.GVA_DB.Model(&row).Updates(map[string]interface{}{
		"status": status, "reviewed_by": reviewerID, "reviewed_at": &now,
	}).Error; err != nil {
		return err
	}
	s.invalidatePlacementCache(row.Placement, row.Audience)
	return nil
}

// @author:		[mo7amed4522](https://github.com/mo7amed4522)
// @function:		BatchReview
// @description:	Approve or reject several recommendations in one transaction
// @param:			ids []uint, reviewerID uint, action string
// @return:		err error
func (s *RecommendationServices) BatchReview(ids []uint, reviewerID uint, action string) (err error) {
	status, err := reviewStatus(action)
	if err != nil {
		return err
	}
	if global.GVA_DB == nil {
		return errors.New("db not init")
	}
	if len(ids) == 0 {
		return errors.New("no recommendations selected")
	}
	var rows []recommendation.Recommendation
	if err := global.GVA_DB.Where("id IN ?", ids).Find(&rows).Error; err != nil {
		return err
	}
	now := time.Now()
	if err := global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		for i := range rows {
			if err := tx.Model(&rows[i]).Updates(map[string]interface{}{
				"status": status, "reviewed_by": reviewerID, "reviewed_at": &now,
			}).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}
	for i := range rows {
		s.invalidatePlacementCache(rows[i].Placement, rows[i].Audience)
	}
	return nil
}

// @author:		[mo7amed4522](https://github.com/mo7amed4522)
// @function:		DeleteRecommendation
// @description:	Soft-delete one recommendation
// @param:			id uint
// @return:		err error
func (s *RecommendationServices) DeleteRecommendation(id uint) (err error) {
	return s.DeleteRecommendationByIds([]uint{id})
}

// @author:		[mo7amed4522](https://github.com/mo7amed4522)
// @function:		DeleteRecommendationByIds
// @description:	Soft-delete several recommendations
// @param:			ids []uint
// @return:		err error
func (s *RecommendationServices) DeleteRecommendationByIds(ids []uint) (err error) {
	if global.GVA_DB == nil {
		return errors.New("db not init")
	}
	if len(ids) == 0 {
		return errors.New("no recommendations selected")
	}
	var rows []recommendation.Recommendation
	if err := global.GVA_DB.Where("id IN ?", ids).Find(&rows).Error; err != nil {
		return err
	}
	if err := global.GVA_DB.Where("id IN ?", ids).Delete(&recommendation.Recommendation{}).Error; err != nil {
		return err
	}
	for i := range rows {
		s.invalidatePlacementCache(rows[i].Placement, rows[i].Audience)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Admin: audit runs + stats
// ---------------------------------------------------------------------------

// @author:		[mo7amed4522](https://github.com/mo7amed4522)
// @function:		GetRunList
// @description:	Paginated audit list of recommendation runs
// @param:			req request.RunSearch
// @return:		list []recommendation.RecommendationRun, total int64, err error
func (s *RecommendationServices) GetRunList(req request.RunSearch) (list []recommendation.RecommendationRun, total int64, err error) {
	if global.GVA_DB == nil {
		return nil, 0, errors.New("db not init")
	}
	db := global.GVA_DB.Model(&recommendation.RecommendationRun{})
	if req.Placement != "" {
		db = db.Where("placement = ?", req.Placement)
	}
	if req.Status != "" {
		db = db.Where("status = ?", req.Status)
	}
	if req.Trigger != "" {
		db = db.Where("trigger = ?", req.Trigger)
	}
	var rows []recommendation.RecommendationRun
	pr, err := utils.PaginateWithCustomQuery(db.Session(&gorm.Session{}), db.Session(&gorm.Session{}).Order("id desc"), req.Page, req.PageSize, &rows)
	if err != nil {
		return nil, 0, err
	}
	return rows, pr.Total, nil
}

// @author:		[mo7amed4522](https://github.com/mo7amed4522)
// @function:		FindRun
// @description:	Get one recommendation run (audit view)
// @param:			id uint
// @return:		run *recommendation.RecommendationRun, err error
func (s *RecommendationServices) FindRun(id uint) (run *recommendation.RecommendationRun, err error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	var row recommendation.RecommendationRun
	if err = global.GVA_DB.First(&row, id).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

// @author:		[mo7amed4522](https://github.com/mo7amed4522)
// @function:		GetStats
// @description:	Impressions, clicks, CTR and bookings per recommendation in a date window
// @param:			placement string, from time.Time, to time.Time
// @return:		stats []response.RecommendationStat, err error
func (s *RecommendationServices) GetStats(placement string, from, to time.Time) (stats []recResp.RecommendationStat, err error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	if to.IsZero() {
		to = time.Now()
	}
	if from.IsZero() {
		from = to.Add(-30 * 24 * time.Hour)
	}
	var rows []struct {
		RecommendationID uint
		Impressions      int64
		Clicks           int64
		Bookings         int64
	}
	q := global.GVA_DB.Table("recommendation_event AS e").
		Select("e.recommendation_id AS recommendation_id, "+
			"SUM(CASE WHEN e.event_type = 'impression' THEN 1 ELSE 0 END) AS impressions, "+
			"SUM(CASE WHEN e.event_type = 'click' THEN 1 ELSE 0 END) AS clicks, "+
			"SUM(CASE WHEN e.event_type = 'booking' THEN 1 ELSE 0 END) AS bookings").
		Where("e.deleted_at IS NULL AND e.created_at >= ? AND e.created_at <= ?", from, to).
		Group("e.recommendation_id")
	if placement != "" {
		q = q.Where("e.placement = ?", placement)
	}
	if err := q.Scan(&rows).Error; err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return []recResp.RecommendationStat{}, nil
	}

	ids := make([]uint, 0, len(rows))
	for i := range rows {
		ids = append(ids, rows[i].RecommendationID)
	}
	var recs []recommendation.Recommendation
	if err := global.GVA_DB.Where("id IN ?", ids).Find(&recs).Error; err != nil {
		return nil, err
	}
	byID := make(map[uint]recommendation.Recommendation, len(recs))
	for i := range recs {
		byID[recs[i].ID] = recs[i]
	}

	stats = make([]recResp.RecommendationStat, 0, len(rows))
	for i := range rows {
		rec := byID[rows[i].RecommendationID]
		stat := recResp.RecommendationStat{
			RecommendationID: rows[i].RecommendationID,
			TargetType:       rec.TargetType, TargetID: rec.TargetID, Placement: rec.Placement,
			Impressions: rows[i].Impressions, Clicks: rows[i].Clicks, Bookings: rows[i].Bookings,
		}
		if stat.Impressions > 0 {
			stat.CTR = float64(stat.Clicks) / float64(stat.Impressions)
		}
		stats = append(stats, stat)
	}
	return stats, nil
}

// @author:		[mo7amed4522](https://github.com/mo7amed4522)
// @function:		TriggerGenerate
// @description:	Admin-triggered generation run for one placement/audience
// @param:			placement string, audience string, adminID uint
// @return:		run *recommendation.RecommendationRun, err error
func (s *RecommendationServices) TriggerGenerate(placement, audience string, adminID uint) (run *recommendation.RecommendationRun, err error) {
	return s.Generate(placement, audience, adminID, "manual")
}

// ParseRange parses a stats window; an empty window means the last 30 days.
func (s *RecommendationServices) ParseRange(from, to string) (time.Time, time.Time, error) {
	var start, end time.Time
	if v, ok := parseQueryTime(from, false); ok {
		start = v
	}
	if v, ok := parseQueryTime(to, true); ok {
		end = v
	}
	if end.IsZero() {
		end = time.Now()
	}
	if start.IsZero() {
		start = end.Add(-30 * 24 * time.Hour)
	}
	if start.After(end) {
		return start, end, errors.New("startTime must be before endTime")
	}
	return start, end, nil
}

// ---------------------------------------------------------------------------
// Public: approved list + event recording
// ---------------------------------------------------------------------------

// @author:		[mo7amed4522](https://github.com/mo7amed4522)
// @function:		GetPublicRecommendations
// @description:	Approved, in-window recommendations whose target is still active; cached
// @param:			placement string, audience string, limit int
// @return:		list []recommendation.Recommendation, err error
func (s *RecommendationServices) GetPublicRecommendations(placement, audience string, limit int) (list []recommendation.Recommendation, err error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	if placement == "" {
		return nil, errors.New("placement is required")
	}
	if audience == "" {
		audience = constants.RecommendationDefaultAudience
	}
	if limit <= 0 {
		limit = defaultMaxItems
	}

	rows, _, err := cache.GetOrLoadJSON(context.Background(), publicCacheKey(placement, audience), publicCacheTTL,
		func() ([]recommendation.Recommendation, error) {
			now := time.Now()
			var cached []recommendation.Recommendation
			if err := global.GVA_DB.
				Where("placement = ? AND audience = ? AND status = ?", placement, audience, constants.RecommendationStatusApproved).
				Where("(starts_at IS NULL OR starts_at <= ?) AND (ends_at IS NULL OR ends_at >= ?)", now, now).
				Order("is_pinned desc, priority desc, score desc").
				Limit(publicPoolLimit).
				Find(&cached).Error; err != nil {
				return nil, err
			}
			return s.onlyLiveTargets(cached)
		})
	if err != nil {
		return nil, err
	}
	if len(rows) > limit {
		rows = rows[:limit]
	}
	return rows, nil
}

// @author:		[mo7amed4522](https://github.com/mo7amed4522)
// @function:		GetPublicRecommendationList
// @description:	Paged customer-facing list, already trimmed to the public response shape
// @param:			placement string, audience string, page int, pageSize int
// @return:		list recResp.PublicRecommendationList, err error
func (s *RecommendationServices) GetPublicRecommendationList(placement, audience string, page, pageSize int) (list recResp.PublicRecommendationList, err error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = defaultMaxItems
	}
	if pageSize > maxPublicPageSize {
		pageSize = maxPublicPageSize
	}
	rows, err := s.GetPublicRecommendations(placement, audience, publicPoolLimit)
	if err != nil {
		return list, err
	}
	items := make([]recResp.PublicRecommendation, 0, len(rows))
	for i := range rows {
		items = append(items, recResp.PublicRecommendation{
			TargetType: rows[i].TargetType, TargetID: rows[i].TargetID, Placement: rows[i].Placement,
			Audience: rows[i].Audience, Score: rows[i].Score, Reason: rows[i].Reason,
			ReasonAr: rows[i].ReasonAr, IsPinned: rows[i].IsPinned,
		})
	}
	total := int64(len(items))
	start := (page - 1) * pageSize
	if start > len(items) {
		start = len(items)
	}
	end := start + pageSize
	if end > len(items) {
		end = len(items)
	}
	return recResp.PublicRecommendationList{Items: items[start:end], Total: total, Page: page, PageSize: pageSize}, nil
}

// onlyLiveTargets drops rows whose target was unpublished, deactivated or deleted since the run.
func (s *RecommendationServices) onlyLiveTargets(rows []recommendation.Recommendation) ([]recommendation.Recommendation, error) {
	if len(rows) == 0 {
		return rows, nil
	}
	pkgIDs := make([]uint, 0, len(rows))
	tripIDs := make([]uint, 0, len(rows))
	agencyIDs := make([]uint, 0, len(rows))
	for _, r := range rows {
		switch r.TargetType {
		case constants.RecommendationTargetTypePackage:
			pkgIDs = append(pkgIDs, r.TargetID)
		case constants.RecommendationTargetTypeTrip:
			tripIDs = append(tripIDs, r.TargetID)
		case constants.RecommendationTargetTypeAgency:
			agencyIDs = append(agencyIDs, r.TargetID)
		}
	}
	live := map[string]bool{}
	if len(pkgIDs) > 0 {
		var ids []uint
		if err := global.GVA_DB.Model(&catalog.Package{}).Where("id IN ? AND status = ?", pkgIDs, constants.PackageStatusActive).Pluck("id", &ids).Error; err != nil {
			return nil, err
		}
		for _, id := range ids {
			live[targetKey(constants.RecommendationTargetTypePackage, id)] = true
		}
	}
	if len(tripIDs) > 0 {
		var ids []uint
		if err := global.GVA_DB.Model(&catalog.CustomTripTemplate{}).Where("id IN ? AND is_active = ?", tripIDs, true).Pluck("id", &ids).Error; err != nil {
			return nil, err
		}
		for _, id := range ids {
			live[targetKey(constants.RecommendationTargetTypeTrip, id)] = true
		}
	}
	if len(agencyIDs) > 0 {
		var ids []uint
		if err := global.GVA_DB.Model(&dxapi.DxAgentProfile{}).Where("id IN ? AND status = ?", agencyIDs, constants.AgentStatusActive).Pluck("id", &ids).Error; err != nil {
			return nil, err
		}
		for _, id := range ids {
			live[targetKey(constants.RecommendationTargetTypeAgency, id)] = true
		}
	}
	out := make([]recommendation.Recommendation, 0, len(rows))
	for _, r := range rows {
		if live[targetKey(r.TargetType, r.TargetID)] {
			out = append(out, r)
		}
	}
	return out, nil
}

// @author:		[mo7amed4522](https://github.com/mo7amed4522)
// @function:		RecordEvent
// @description:	Store an impression/click/booking signal (rate limited, no PII)
// @param:			req request.RecommendationEventReq, userID uint
// @return:		err error
func (s *RecommendationServices) RecordEvent(req request.RecommendationEventReq, userID uint) (err error) {
	if global.GVA_DB == nil {
		return errors.New("db not init")
	}
	switch req.EventType {
	case "impression", "click", "booking":
	default:
		return errors.New("invalid event type")
	}
	var rec recommendation.Recommendation
	if err := global.GVA_DB.First(&rec, req.RecommendationID).Error; err != nil {
		return errors.New("recommendation not found")
	}
	if rec.Status != constants.RecommendationStatusApproved {
		return errors.New("recommendation is not published")
	}
	placement := req.Placement
	if placement == "" {
		placement = rec.Placement
	}
	var by *uint
	if userID > 0 {
		uid := userID
		by = &uid
	}
	return global.GVA_DB.Create(&recommendation.RecommendationEvent{
		RecommendationID: rec.ID, EventType: req.EventType, UserID: by, Placement: placement,
	}).Error
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// ensureTarget validates that the referenced item exists and is currently active.
func (s *RecommendationServices) ensureTarget(targetType string, targetID uint) error {
	if targetID == 0 {
		return errors.New("targetId is required")
	}
	var count int64
	var err error
	switch targetType {
	case constants.RecommendationTargetTypePackage:
		err = global.GVA_DB.Model(&catalog.Package{}).Where("id = ? AND status = ?", targetID, constants.PackageStatusActive).Count(&count).Error
	case constants.RecommendationTargetTypeTrip:
		err = global.GVA_DB.Model(&catalog.CustomTripTemplate{}).Where("id = ? AND is_active = ?", targetID, true).Count(&count).Error
	case constants.RecommendationTargetTypeAgency:
		err = global.GVA_DB.Model(&dxapi.DxAgentProfile{}).Where("id = ? AND status = ?", targetID, constants.AgentStatusActive).Count(&count).Error
	default:
		return fmt.Errorf("unsupported target type %q", targetType)
	}
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("%s %d was not found or is not active", targetType, targetID)
	}
	return nil
}

// invalidatePlacementCache clears the public cache keys for a placement (all known audiences).
func (s *RecommendationServices) invalidatePlacementCache(placement, audience string) {
	if placement == "" || global.GVA_DB == nil {
		return
	}
	audiences := map[string]struct{}{constants.RecommendationDefaultAudience: {}, audience: {}}
	if audience == "" {
		delete(audiences, "")
	}
	var known []string
	if err := global.GVA_DB.Model(&recommendation.Recommendation{}).
		Where("placement = ?", placement).Distinct().Pluck("audience", &known).Error; err == nil {
		for _, a := range known {
			if a != "" {
				audiences[a] = struct{}{}
			}
		}
	}
	for a := range audiences {
		key := publicCacheKey(placement, a)
		_ = cache.DeleteRedis(context.Background(), key)
		cache.Delete(key)
	}
}

func publicCacheKey(placement, audience string) string {
	return fmt.Sprintf("recommendation:%s:%s", placement, audience)
}

func targetKey(targetType string, targetID uint) string {
	return fmt.Sprintf("%s:%d", targetType, targetID)
}

func reviewStatus(action string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(action)) {
	case "approve":
		return constants.RecommendationStatusApproved, nil
	case "reject":
		return constants.RecommendationStatusRejected, nil
	default:
		return "", fmt.Errorf("invalid review action %q", action)
	}
}

func clampScore(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func truncateReason(v string) string {
	v = strings.TrimSpace(v)
	r := []rune(v)
	if len(r) > maxReasonLength {
		return strings.TrimSpace(string(r[:maxReasonLength]))
	}
	return v
}

// parseQueryTime accepts YYYY-MM-DD, "YYYY-MM-DD HH:MM:SS" and RFC3339. endOfDay expands a
// date-only value to the end of that day so ranges are inclusive.
func parseQueryTime(v string, endOfDay bool) (time.Time, bool) {
	v = strings.TrimSpace(v)
	if v == "" {
		return time.Time{}, false
	}
	for _, layout := range []string{"2006-01-02", "2006-01-02 15:04:05", time.RFC3339} {
		t, err := time.ParseInLocation(layout, v, time.Local)
		if err != nil {
			continue
		}
		if endOfDay && layout == "2006-01-02" {
			t = t.Add(24*time.Hour - time.Second)
		}
		return t, true
	}
	return time.Time{}, false
}

// logDrop records rejected LLM ids without echoing the whole prompt.
func logDropped(run *recommendation.RecommendationRun, dropped []string) {
	if len(dropped) == 0 {
		return
	}
	if global.GVA_LOG != nil {
		global.GVA_LOG.Warn("recommendation: dropped ids not present in the candidate set",
			zap.Uint("runId", run.ID), zap.Strings("dropped", dropped))
	}
}
