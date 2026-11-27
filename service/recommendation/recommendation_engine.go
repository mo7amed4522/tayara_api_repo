package recommendation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/mo7amed4522/tayara_api/constants"
	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/model/catalog"
	"github.com/mo7amed4522/tayara_api/model/recommendation"
	"github.com/mo7amed4522/tayara_api/utils"
	"github.com/mo7amed4522/tayara_api/utils/llm"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// @author:		[mo7amed4522](https://github.com/mo7amed4522)
// @function:		Generate
// @description:	Build candidates, ask the LLM, validate strictly, persist (rules fallback on failure)
// @param:			placement string, audience string, triggeredBy uint, trigger string
// @return:		run *recommendation.RecommendationRun, err error
func (s *RecommendationServices) Generate(placement, audience string, triggeredBy uint, trigger string) (run *recommendation.RecommendationRun, err error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	if placement == "" {
		return nil, errors.New("placement is required")
	}
	if audience == "" {
		audience = constants.RecommendationDefaultAudience
	}
	if trigger != "manual" {
		trigger = "cron"
	}

	release, ok := utils.TryLock(context.Background(), "recommendation:generate:"+placement+":"+audience, 5*time.Minute)
	if !ok {
		return nil, fmt.Errorf("a generation run for %s/%s is already in progress", placement, audience)
	}
	defer release()

	setting, err := s.findSetting(placement)
	if err != nil {
		return nil, err
	}
	if !setting.Enabled {
		return nil, fmt.Errorf("recommendation generation is disabled for placement %q", placement)
	}

	started := time.Now()
	cfg := global.GVA_CONFIG.LLM
	run = &recommendation.RecommendationRun{
		Trigger: trigger, TriggeredBy: triggeredBy, Placement: placement, Audience: audience,
		Provider: cfg.Provider, ModelName: cfg.Model, PromptVersion: setting.PromptVersion,
		Status: constants.RecommendationRunStatusRunning, StartedAt: &started,
	}
	if err := global.GVA_DB.Create(run).Error; err != nil {
		return nil, err
	}

	cands, err := s.buildCandidates(setting)
	if err != nil {
		run.Status = constants.RecommendationRunStatusFailed
		run.ErrorMessage = err.Error()
		s.finishRun(run)
		return run, err
	}
	run.CandidateCount = len(cands)
	if snapshot, mErr := json.Marshal(cands); mErr == nil {
		run.InputSnapshot = string(snapshot)
	}

	items, raw, llmErr := s.askLLM(setting, placement, audience, cands)
	run.RawOutput = raw
	source := constants.RecommendationSourceLLM
	if llmErr != nil {
		// Graceful fallback: customers must never see an error because the LLM is down.
		items = rankByRules(cands, setting.MaxItems)
		source = constants.RecommendationSourceRules
		run.Status = constants.RecommendationRunStatusFallback
		run.ErrorMessage = llmErr.Error()
	} else {
		valid, dropped := validateItems(items, cands, setting.MaxItems)
		logDropped(run, dropped)
		items = valid
		run.Status = constants.RecommendationRunStatusSuccess
		if len(items) == 0 {
			items = rankByRules(cands, setting.MaxItems)
			source = constants.RecommendationSourceRules
			run.Status = constants.RecommendationRunStatusFallback
			run.ErrorMessage = "llm returned no valid candidates"
		}
	}

	persisted, persistErr := s.persistItems(run, setting, items, source)
	if persistErr != nil {
		run.Status = constants.RecommendationRunStatusFailed
		run.ErrorMessage = persistErr.Error()
		s.finishRun(run)
		return run, persistErr
	}
	run.ResultCount = persisted
	s.finishRun(run)
	s.invalidatePlacementCache(placement, audience)

	if global.GVA_LOG != nil {
		global.GVA_LOG.Info("recommendation run finished",
			zap.Uint("runId", run.ID), zap.String("placement", placement), zap.String("audience", audience),
			zap.String("status", run.Status), zap.Int("candidates", run.CandidateCount), zap.Int("persisted", persisted))
	}
	return run, nil
}

// @author:		[mo7amed4522](https://github.com/mo7amed4522)
// @function:		GenerateDue
// @description:	Cron entry point: regenerate each enabled placement whose refresh interval elapsed
// @return:		err error
func (s *RecommendationServices) GenerateDue() (err error) {
	if global.GVA_DB == nil {
		return errors.New("db not init")
	}
	var settings []recommendation.RecommendationSetting
	if err := global.GVA_DB.Where("enabled = ?", true).Find(&settings).Error; err != nil {
		return err
	}
	for i := range settings {
		interval := parseRefreshInterval(settings[i].RefreshCron)
		var last recommendation.RecommendationRun
		lastErr := global.GVA_DB.
			Where("placement = ? AND audience = ? AND status IN ?", settings[i].Placement, constants.RecommendationDefaultAudience, []string{constants.RecommendationRunStatusSuccess, constants.RecommendationRunStatusFallback}).
			Order("id desc").First(&last).Error
		if lastErr == nil && last.StartedAt != nil && time.Since(*last.StartedAt) < interval {
			continue
		}
		if _, genErr := s.Generate(settings[i].Placement, constants.RecommendationDefaultAudience, 0, "cron"); genErr != nil {
			if global.GVA_LOG != nil {
				global.GVA_LOG.Warn("recommendation cron run failed",
					zap.String("placement", settings[i].Placement), zap.Error(genErr))
			}
		}
	}
	return nil
}

// buildCandidates reads only active/published items, capped at the setting's candidate limit.
func (s *RecommendationServices) buildCandidates(setting *recommendation.RecommendationSetting) ([]recommendation.RecommendationCandidate, error) {
	limit := setting.CandidateLimit
	if limit <= 0 {
		limit = 50
	}
	allowed := allowedTargetTypes(setting)
	out := make([]recommendation.RecommendationCandidate, 0, limit)

	// Rough allocation so one type never crowds out the others.
	typeCaps := map[string]int{
		constants.RecommendationTargetTypePackage: limit * 60 / 100,
		constants.RecommendationTargetTypeTrip:    limit * 20 / 100,
		constants.RecommendationTargetTypeAgency:  limit * 20 / 100,
	}
	for k, v := range typeCaps {
		if v < 1 {
			typeCaps[k] = 1
		}
	}

	if allowed[constants.RecommendationTargetTypePackage] {
		rows, err := s.packageCandidates(typeCaps[constants.RecommendationTargetTypePackage])
		if err != nil {
			return nil, err
		}
		out = append(out, rows...)
	}
	if allowed[constants.RecommendationTargetTypeTrip] {
		rows, err := s.tripCandidates(typeCaps[constants.RecommendationTargetTypeTrip])
		if err != nil {
			return nil, err
		}
		out = append(out, rows...)
	}
	if allowed[constants.RecommendationTargetTypeAgency] {
		rows, err := s.agencyCandidates(typeCaps[constants.RecommendationTargetTypeAgency])
		if err != nil {
			return nil, err
		}
		out = append(out, rows...)
	}

	s.attachSignals(out)
	if len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func (s *RecommendationServices) packageCandidates(limit int) ([]recommendation.RecommendationCandidate, error) {
	if limit <= 0 {
		return nil, nil
	}
	agentRatings := global.GVA_DB.Table("engage_review").
		Select("agent_id, AVG(rating) AS avg_rating").
		Where("review_status = ? AND agent_id IS NOT NULL AND deleted_at IS NULL", constants.ReviewStatusActive).
		Group("agent_id")

	var rows []struct {
		ID           uint
		Name         string
		Category     string
		Location     string
		Destination  string
		DisplayPrice int64
		CreatedAt    time.Time
		AgentRating  float64
	}
	if err := global.GVA_DB.Table("catalog_package AS p").
		Select("p.id, p.name, p.category, p.location, p.destination, p.display_price, p.created_at, COALESCE(ar.avg_rating, 0) AS agent_rating").
		Joins("LEFT JOIN (?) AS ar ON ar.agent_id = p.agent_id", agentRatings).
		Where("p.status = ? AND p.deleted_at IS NULL", constants.PackageStatusActive).
		Order("p.created_at desc").
		Limit(limit).
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}

	ids := make([]uint, 0, len(rows))
	for _, r := range rows {
		ids = append(ids, r.ID)
	}
	rating := map[uint]float64{}
	reviews := map[uint]int64{}
	var ratingRows []struct {
		PackageID   uint
		AvgRating   float64
		ReviewCount int64
	}
	if err := global.GVA_DB.Table("engage_review").
		Select("package_id, AVG(rating) AS avg_rating, COUNT(*) AS review_count").
		Where("review_status = ? AND package_id IN ? AND deleted_at IS NULL", constants.ReviewStatusActive, ids).
		Group("package_id").Scan(&ratingRows).Error; err == nil {
		for _, r := range ratingRows {
			rating[r.PackageID] = r.AvgRating
			reviews[r.PackageID] = r.ReviewCount
		}
	}
	bookings := map[uint]int64{}
	var bookingRows []struct {
		PackageID uint
		Total     int64
	}
	since := time.Now().Add(-30 * 24 * time.Hour)
	if err := global.GVA_DB.Table("bookings").
		Select("package_id, COUNT(*) AS total").
		Where("created_at >= ? AND package_id IN ?", since, ids).
		Group("package_id").Scan(&bookingRows).Error; err == nil {
		for _, r := range bookingRows {
			bookings[r.PackageID] = r.Total
		}
	}

	out := make([]recommendation.RecommendationCandidate, 0, len(rows))
	for _, r := range rows {
		out = append(out, recommendation.RecommendationCandidate{
			TargetType: constants.RecommendationTargetTypePackage, TargetID: r.ID, Title: r.Name,
			Category: r.Category, Destination: firstNonEmpty(r.Location, r.Destination),
			PriceBand: priceBand(r.DisplayPrice), Rating: rating[r.ID], ReviewCount: reviews[r.ID],
			Bookings30d: bookings[r.ID], RecencyDays: daysSince(r.CreatedAt), AgencyRating: r.AgentRating,
		})
	}
	return out, nil
}

func (s *RecommendationServices) tripCandidates(limit int) ([]recommendation.RecommendationCandidate, error) {
	if limit <= 0 {
		return nil, nil
	}
	var rows []catalog.CustomTripTemplate
	if err := global.GVA_DB.
		Where("is_active = ?", true).
		Order("created_at desc").
		Limit(limit).
		Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]recommendation.RecommendationCandidate, 0, len(rows))
	for i := range rows {
		out = append(out, recommendation.RecommendationCandidate{
			TargetType: constants.RecommendationTargetTypeTrip, TargetID: rows[i].ID, Title: rows[i].Name,
			Category: "custom_trip", RecencyDays: daysSince(rows[i].CreatedAt),
		})
	}
	return out, nil
}

func (s *RecommendationServices) agencyCandidates(limit int) ([]recommendation.RecommendationCandidate, error) {
	if limit <= 0 {
		return nil, nil
	}
	ratings := global.GVA_DB.Table("engage_review").
		Select("agent_id, AVG(rating) AS avg_rating, COUNT(*) AS review_count").
		Where("review_status = ? AND agent_id IS NOT NULL AND deleted_at IS NULL", constants.ReviewStatusActive).
		Group("agent_id")

	var rows []struct {
		ID          uint
		CompanyName string
		Country     string
		CreatedAt   time.Time
		AvgRating   float64
		ReviewCount int64
	}
	if err := global.GVA_DB.Table("dx_agent_profile AS a").
		Select("a.id, a.company_name, a.country, a.created_at, COALESCE(r.avg_rating, 0) AS avg_rating, COALESCE(r.review_count, 0) AS review_count").
		Joins("LEFT JOIN (?) AS r ON r.agent_id = a.id", ratings).
		Where("a.status = ? AND a.deleted_at IS NULL", constants.AgentStatusActive).
		Order("a.created_at desc").
		Limit(limit).
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]recommendation.RecommendationCandidate, 0, len(rows))
	for _, r := range rows {
		out = append(out, recommendation.RecommendationCandidate{
			TargetType: constants.RecommendationTargetTypeAgency, TargetID: r.ID, Title: r.CompanyName,
			Destination: r.Country, Rating: r.AvgRating, ReviewCount: r.ReviewCount,
			RecencyDays: daysSince(r.CreatedAt),
		})
	}
	return out, nil
}

// attachSignals adds aggregated impression/click/booking counts (never per-user data).
func (s *RecommendationServices) attachSignals(cands []recommendation.RecommendationCandidate) {
	if len(cands) == 0 {
		return
	}
	signals := make([]recommendation.RecommendationSignal, 0, len(cands))
	if err := global.GVA_DB.Table("recommendation_event AS e").
		Select("r.target_type AS target_type, r.target_id AS target_id, "+
			"SUM(CASE WHEN e.event_type = 'impression' THEN 1 ELSE 0 END) AS impressions, "+
			"SUM(CASE WHEN e.event_type = 'click' THEN 1 ELSE 0 END) AS clicks, "+
			"SUM(CASE WHEN e.event_type = 'booking' THEN 1 ELSE 0 END) AS bookings").
		Joins("JOIN recommendation AS r ON r.id = e.recommendation_id").
		Where("e.deleted_at IS NULL AND r.deleted_at IS NULL AND e.created_at >= ?", time.Now().Add(-90*24*time.Hour)).
		Group("r.target_type, r.target_id").
		Scan(&signals).Error; err != nil {
		return
	}
	index := make(map[string]int, len(cands))
	for i := range cands {
		index[targetKey(cands[i].TargetType, cands[i].TargetID)] = i
	}
	for i := range signals {
		if idx, found := index[targetKey(signals[i].TargetType, signals[i].TargetID)]; found {
			cands[idx].Impressions = signals[i].Impressions
			cands[idx].Clicks = signals[i].Clicks
			cands[idx].Bookings = signals[i].Bookings
		}
	}
}

// askLLM performs the single provider call behind the Client interface, with a hard timeout.
func (s *RecommendationServices) askLLM(setting *recommendation.RecommendationSetting, placement, audience string, cands []recommendation.RecommendationCandidate) ([]recommendation.RecommendationItem, string, error) {
	cfg := global.GVA_CONFIG.LLM
	client, err := llm.NewClient(llm.ConfigFrom(cfg.Provider, cfg.Model, cfg.APIKey, cfg.BaseURL,
		cfg.TimeoutSeconds, cfg.MaxRetries, cfg.Temperature))
	if err != nil {
		return nil, "", err
	}
	payload, err := json.Marshal(map[string]interface{}{
		"placement": placement, "audience": audience, "candidates": cands,
	})
	if err != nil {
		return nil, "", err
	}
	timeout := cfg.TimeoutSeconds
	if timeout <= 0 {
		timeout = 20
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	raw, _, err := client.Complete(ctx, renderPrompt(setting.PromptTemplate, setting.MaxItems), string(payload))
	if err != nil {
		return nil, raw, err
	}
	items, parseErr := parseLLMItems(raw)
	if parseErr != nil {
		return nil, raw, parseErr
	}
	return items, raw, nil
}

// parseLLMItems tolerates code fences/prose around the JSON object but nothing else.
func parseLLMItems(raw string) ([]recommendation.RecommendationItem, error) {
	body := strings.TrimSpace(raw)
	body = strings.TrimPrefix(body, "```json")
	body = strings.TrimPrefix(body, "```")
	body = strings.TrimSuffix(body, "```")
	if start := strings.Index(body, "{"); start >= 0 {
		if end := strings.LastIndex(body, "}"); end > start {
			body = body[start : end+1]
		}
	}
	var parsed struct {
		Recommendations []recommendation.RecommendationItem `json:"recommendations"`
	}
	if err := json.Unmarshal([]byte(body), &parsed); err != nil {
		return nil, fmt.Errorf("invalid llm json: %w", err)
	}
	return parsed.Recommendations, nil
}

// validateItems enforces grounding: only candidate ids survive, scores are clamped, reasons are
// truncated, duplicates are removed and the same agency may appear at most twice.
func validateItems(items []recommendation.RecommendationItem, cands []recommendation.RecommendationCandidate, maxItems int) (valid []recommendation.RecommendationItem, dropped []string) {
	known := make(map[string]struct{}, len(cands))
	for _, c := range cands {
		known[targetKey(c.TargetType, c.TargetID)] = struct{}{}
	}
	if maxItems <= 0 {
		maxItems = defaultMaxItems
	}
	seen := make(map[string]struct{}, len(items))
	agencyCount := 0
	valid = make([]recommendation.RecommendationItem, 0, len(items))
	for _, it := range items {
		it.TargetType = strings.ToLower(strings.TrimSpace(it.TargetType))
		key := targetKey(it.TargetType, it.TargetID)
		if _, ok := known[key]; !ok {
			dropped = append(dropped, key)
			continue
		}
		if _, dup := seen[key]; dup {
			continue
		}
		if it.TargetType == constants.RecommendationTargetTypeAgency {
			if agencyCount >= 2 {
				dropped = append(dropped, key+" (agency cap)")
				continue
			}
			agencyCount++
		}
		seen[key] = struct{}{}
		it.Score = clampScore(it.Score)
		it.Reason = truncateReason(it.Reason)
		valid = append(valid, it)
		if len(valid) >= maxItems {
			break
		}
	}
	return valid, dropped
}

// rankByRules the deterministic fallback: rating, popularity and recency.
func rankByRules(cands []recommendation.RecommendationCandidate, maxItems int) []recommendation.RecommendationItem {
	if maxItems <= 0 {
		maxItems = defaultMaxItems
	}
	type scored struct {
		item  recommendation.RecommendationItem
		score float64
	}
	ranked := make([]scored, 0, len(cands))
	for _, c := range cands {
		rating := c.Rating
		if rating == 0 {
			rating = c.AgencyRating
		}
		ratingNorm := rating / 5
		if ratingNorm > 1 {
			ratingNorm = 1
		}
		popularity := float64(c.Bookings30d) / 20
		if popularity > 1 {
			popularity = 1
		}
		recency := 1 / (1 + float64(c.RecencyDays)/30)
		engagement := float64(c.Clicks+c.Bookings*3) / 50
		if engagement > 1 {
			engagement = 1
		}
		score := 0.40*ratingNorm + 0.25*popularity + 0.20*recency + 0.15*engagement
		ranked = append(ranked, scored{
			item: recommendation.RecommendationItem{
				TargetType: c.TargetType, TargetID: c.TargetID,
				Score:  clampScore(score),
				Reason: truncateReason(ruleReason(c)),
			},
			score: score,
		})
	}
	sort.SliceStable(ranked, func(i, j int) bool { return ranked[i].score > ranked[j].score })
	if len(ranked) > maxItems {
		ranked = ranked[:maxItems]
	}
	out := make([]recommendation.RecommendationItem, 0, len(ranked))
	for _, r := range ranked {
		out = append(out, r.item)
	}
	return out
}

func ruleReason(c recommendation.RecommendationCandidate) string {
	switch {
	case c.ReviewCount > 0 && c.Bookings30d > 0:
		return "Highly rated and frequently booked in the last 30 days."
	case c.ReviewCount > 0:
		return "Strong customer rating with recent reviews."
	case c.Bookings30d > 0:
		return "Popular with recent bookings."
	default:
		return "Newly published item matching this placement."
	}
}

// persistItems writes validated items in one transaction. Pinned and admin-curated rows are never
// overwritten, and rows no longer returned by the run are expired.
func (s *RecommendationServices) persistItems(run *recommendation.RecommendationRun, setting *recommendation.RecommendationSetting, items []recommendation.RecommendationItem, source string) (int, error) {
	if global.GVA_DB == nil {
		return 0, errors.New("db not init")
	}
	persisted := 0
	err := global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		var existing []recommendation.Recommendation
		if err := tx.Where("placement = ? AND audience = ?", run.Placement, run.Audience).Find(&existing).Error; err != nil {
			return err
		}
		byKey := make(map[string]*recommendation.Recommendation, len(existing))
		for i := range existing {
			byKey[targetKey(existing[i].TargetType, existing[i].TargetID)] = &existing[i]
		}

		seen := make(map[string]struct{}, len(items))
		for _, item := range items {
			key := targetKey(item.TargetType, item.TargetID)
			if _, dup := seen[key]; dup {
				continue
			}
			seen[key] = struct{}{}

			status := constants.RecommendationStatusPending
			if setting.AutoApprove && item.Score >= setting.AutoApproveThreshold {
				status = constants.RecommendationStatusApproved
			}

			if row, found := byKey[key]; found {
				if row.IsPinned || row.Source == constants.RecommendationSourceAdmin || row.Status == constants.RecommendationStatusRejected {
					continue
				}
				if row.Status == constants.RecommendationStatusApproved {
					status = constants.RecommendationStatusApproved
				}
				if err := tx.Model(row).Updates(map[string]interface{}{
					"score": item.Score, "reason": item.Reason, "run_id": run.ID,
					"source": source, "status": status,
				}).Error; err != nil {
					return err
				}
				persisted++
				continue
			}

			// A soft-deleted row still occupies the unique key; revive it instead of duplicating.
			var dead recommendation.Recommendation
			if err := tx.Unscoped().
				Where("target_type = ? AND target_id = ? AND placement = ? AND audience = ? AND deleted_at IS NOT NULL",
					item.TargetType, item.TargetID, run.Placement, run.Audience).
				First(&dead).Error; err == nil {
				if err := tx.Unscoped().Model(&dead).Updates(map[string]interface{}{
					"deleted_at": nil, "score": item.Score, "reason": item.Reason, "run_id": run.ID,
					"source": source, "status": status, "is_pinned": false,
				}).Error; err != nil {
					return err
				}
				persisted++
				continue
			}

			rec := recommendation.Recommendation{
				TargetType: item.TargetType, TargetID: item.TargetID,
				Placement: run.Placement, Audience: run.Audience,
				Source: source, Status: status, Score: item.Score, Reason: item.Reason, RunID: run.ID,
			}
			if err := tx.Create(&rec).Error; err != nil {
				return err
			}
			persisted++
		}

		for i := range existing {
			row := &existing[i]
			if row.IsPinned || row.Source == constants.RecommendationSourceAdmin || row.Status == constants.RecommendationStatusRejected {
				continue
			}
			if row.Source != constants.RecommendationSourceLLM && row.Source != constants.RecommendationSourceRules {
				continue
			}
			if _, returned := seen[targetKey(row.TargetType, row.TargetID)]; returned {
				continue
			}
			if err := tx.Model(row).Update("status", constants.RecommendationStatusExpired).Error; err != nil {
				return err
			}
		}
		return nil
	})
	return persisted, err
}

// finishRun persists the audit facts and duration.
func (s *RecommendationServices) finishRun(run *recommendation.RecommendationRun) {
	now := time.Now()
	run.FinishedAt = &now
	if run.StartedAt != nil {
		run.DurationMs = now.Sub(*run.StartedAt).Milliseconds()
	}
	if global.GVA_DB == nil {
		return
	}
	if err := global.GVA_DB.Model(run).Updates(map[string]interface{}{
		"status": run.Status, "error_message": run.ErrorMessage, "raw_output": run.RawOutput,
		"input_snapshot": run.InputSnapshot, "candidate_count": run.CandidateCount,
		"result_count": run.ResultCount, "duration_ms": run.DurationMs, "finished_at": run.FinishedAt,
		"provider": run.Provider, "model_name": run.ModelName, "prompt_version": run.PromptVersion,
	}).Error; err != nil && global.GVA_LOG != nil {
		global.GVA_LOG.Warn("recommendation: could not persist run result", zap.Uint("runId", run.ID), zap.Error(err))
	}
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func renderPrompt(template string, maxItems int) string {
	if strings.TrimSpace(template) == "" {
		template = DefaultPromptTemplate
	}
	if maxItems <= 0 {
		maxItems = defaultMaxItems
	}
	return strings.ReplaceAll(template, "{{max_items}}", strconv.Itoa(maxItems))
}

func allowedTargetTypes(setting *recommendation.RecommendationSetting) map[string]bool {
	out := map[string]bool{constants.RecommendationTargetTypePackage: true, constants.RecommendationTargetTypeTrip: true, constants.RecommendationTargetTypeAgency: true}
	if len(setting.TargetTypes) == 0 {
		return out
	}
	chosen := map[string]bool{}
	collect := func(v interface{}) {
		switch list := v.(type) {
		case []interface{}:
			for _, raw := range list {
				if s, ok := raw.(string); ok {
					chosen[strings.ToLower(strings.TrimSpace(s))] = true
				}
			}
		case []string:
			for _, s := range list {
				chosen[strings.ToLower(strings.TrimSpace(s))] = true
			}
		}
	}
	if list, ok := setting.TargetTypes["targetTypes"]; ok {
		collect(list)
	}
	if list, ok := setting.TargetTypes["items"]; ok {
		collect(list)
	}
	if list, ok := setting.TargetTypes["types"]; ok {
		collect(list)
	}
	for k, v := range setting.TargetTypes {
		for _, t := range []string{constants.RecommendationTargetTypePackage, constants.RecommendationTargetTypeTrip, constants.RecommendationTargetTypeAgency} {
			if strings.EqualFold(k, t) {
				if b, isBool := v.(bool); isBool && b {
					chosen[t] = true
				}
			}
		}
	}
	if len(chosen) == 0 {
		return out
	}
	filtered := map[string]bool{}
	for t := range chosen {
		if _, ok := out[t]; ok {
			filtered[t] = true
		}
	}
	if len(filtered) == 0 {
		return out
	}
	return filtered
}

func priceBand(minorUnits int64) string {
	switch {
	case minorUnits <= 0:
		return ""
	case minorUnits < 300000:
		return "budget"
	case minorUnits < 800000:
		return "mid"
	default:
		return "premium"
	}
}

func daysSince(t time.Time) int {
	if t.IsZero() {
		return 365
	}
	d := int(time.Since(t).Hours() / 24)
	if d < 0 {
		return 0
	}
	return d
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

// parseRefreshInterval accepts 6h / 15m / @every 6h / "0 */6 * * *" and defaults to 6h.
func parseRefreshInterval(v string) time.Duration {
	v = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(v), "@every"))
	if v == "" {
		return 6 * time.Hour
	}
	if d, err := time.ParseDuration(v); err == nil && d > 0 {
		return d
	}
	fields := strings.Fields(v)
	if len(fields) == 5 {
		hourField := fields[1]
		if strings.HasPrefix(hourField, "*/") {
			var n int
			if _, err := fmt.Sscanf(hourField, "*/%d", &n); err == nil && n > 0 {
				return time.Duration(n) * time.Hour
			}
		}
	}
	return 6 * time.Hour
}
