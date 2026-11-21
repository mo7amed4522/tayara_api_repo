package engage

import (
	"errors"
	"time"

	"github.com/mo7amed4522/tayara_api/constants"
	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/model/engage"
	engageReq "github.com/mo7amed4522/tayara_api/model/engage/request"
	engageResp "github.com/mo7amed4522/tayara_api/model/engage/response"
	"github.com/mo7amed4522/tayara_api/utils"
	"gorm.io/gorm"
)

const maxAdsPerPlacement = 3

type AdServices struct{}

// CreateAd creates an ad (draft). Enforces the max-3-per-placement rule and a unique display
// order within the placement.
func (a AdServices) CreateAd(adminID uint, req engageReq.AdCreativeReq) (*engage.AdCreative, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	if err := a.validatePlacement(req); err != nil {
		return nil, err
	}
	ad := &engage.AdCreative{
		AdCode: utils.NextCode(constants.CodePrefixAd), Title: req.Title, Placement: req.Placement,
		Status: constants.AdStatusDraft, DisplayOrder: req.DisplayOrder,
		StartDatetime: req.StartDatetime, EndDatetime: req.EndDatetime, BannerImageURL: req.BannerImageURL,
		OverlayTitle: req.OverlayTitle, OverlaySubtitle: req.OverlaySubtitle, CtaLabel: req.CtaLabel,
		RedirectType: req.RedirectType, TargetCountryID: req.TargetCountryID,
		TargetAgentID: req.TargetAgentID, TargetPackageID: req.TargetPackageID, CreatedBy: adminID,
	}
	if err := global.GVA_DB.Create(ad).Error; err != nil {
		return nil, err
	}
	return ad, nil
}

// UpdateAd updates an ad (only non-terminal states).
func (a AdServices) UpdateAd(adminID uint, req engageReq.AdCreativeReq) (*engage.AdCreative, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	var ad engage.AdCreative
	if err := global.GVA_DB.First(&ad, req.ID).Error; err != nil {
		return nil, errors.New("ad not found")
	}
	if ad.Status == constants.AdStatusExpired || ad.Status == constants.AdStatusActive {
		return nil, errors.New("expired/active ads cannot be edited")
	}
	if err := global.GVA_DB.Model(&engage.AdCreative{}).Where("id = ?", req.ID).Updates(map[string]interface{}{
		"title": req.Title, "placement": req.Placement, "display_order": req.DisplayOrder,
		"start_datetime": req.StartDatetime, "end_datetime": req.EndDatetime,
		"banner_image_url": req.BannerImageURL, "overlay_title": req.OverlayTitle,
		"overlay_subtitle": req.OverlaySubtitle, "cta_label": req.CtaLabel, "redirect_type": req.RedirectType,
		"target_country_id": req.TargetCountryID, "target_agent_id": req.TargetAgentID, "target_package_id": req.TargetPackageID,
	}).Error; err != nil {
		return nil, err
	}
	return a.GetAd(req.ID)
}

// SetStatus transitions an ad to active/disabled/scheduled. Expired is terminal.
func (a AdServices) SetStatus(req engageReq.AdSetStatusReq) error {
	if global.GVA_DB == nil {
		return errors.New("db not init")
	}
	var ad engage.AdCreative
	if err := global.GVA_DB.First(&ad, req.ID).Error; err != nil {
		return errors.New("ad not found")
	}
	if ad.Status == constants.AdStatusExpired {
		return errors.New("expired ads cannot change status")
	}
	switch req.Status {
	case constants.AdStatusActive:
		// must have a banner + not exceed placement cap
		if ad.BannerImageURL == "" {
			return errors.New("ad needs a banner image before activation")
		}
		var activeCount int64
		_ = global.GVA_DB.Model(&engage.AdCreative{}).
			Where("placement = ? AND status IN ?", ad.Placement, []string{constants.AdStatusActive, constants.AdStatusScheduled}).
			Count(&activeCount).Error
		if activeCount >= maxAdsPerPlacement {
			return errors.New("placement already has the maximum active/scheduled ads")
		}
	case constants.AdStatusDisabled, constants.AdStatusScheduled, constants.AdStatusDraft:
		// allowed
	default:
		return errors.New("unsupported ad status")
	}
	return global.GVA_DB.Model(&engage.AdCreative{}).Where("id = ?", req.ID).
		Update("status", req.Status).Error
}

// DeleteAd soft-deletes an ad (draft/disabled only).
func (a AdServices) DeleteAd(id uint) error {
	if global.GVA_DB == nil {
		return errors.New("db not init")
	}
	var ad engage.AdCreative
	if err := global.GVA_DB.First(&ad, id).Error; err != nil {
		return errors.New("ad not found")
	}
	if ad.Status == constants.AdStatusActive || ad.Status == constants.AdStatusExpired {
		return errors.New("active/expired ads cannot be deleted")
	}
	return global.GVA_DB.Delete(&engage.AdCreative{}, id).Error
}

// RecordImpression bumps the impression counter for an ad.
func (a AdServices) RecordImpression(adID uint) error {
	if global.GVA_DB == nil {
		return errors.New("db not init")
	}
	return global.GVA_DB.Model(&engage.AdCreative{}).Where("id = ?", adID).
		UpdateColumn("impressions", gorm.Expr("impressions + ?", 1)).Error
}

// RecordClick records a click event and increments the creative's click counter.
func (a AdServices) RecordClick(userID *uint, req engageReq.AdClickReq) (*engage.AdClick, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	var ad engage.AdCreative
	if err := global.GVA_DB.First(&ad, req.AdID).Error; err != nil {
		return nil, errors.New("ad not found")
	}
	click := &engage.AdClick{AdID: req.AdID, UserID: userID, SourcePage: req.SourcePage}
	if err := global.GVA_DB.Create(click).Error; err != nil {
		return nil, err
	}
	_ = global.GVA_DB.Model(&engage.AdCreative{}).Where("id = ?", req.AdID).
		UpdateColumn("clicks", gorm.Expr("clicks + ?", 1)).Error
	return click, nil
}

// ListAds returns ads filtered by placement/status (admin) with computed CTR.
func (a AdServices) ListAds(req engageReq.AdCreativePageReq) ([]engageResp.AdDetail, int64, int64, error) {
	if global.GVA_DB == nil {
		return nil, 0, 0, errors.New("db not init")
	}
	db := global.GVA_DB.Model(&engage.AdCreative{})
	if req.Placement != "" {
		db = db.Where("placement = ?", req.Placement)
	}
	if req.Status != "" {
		db = db.Where("status = ?", req.Status)
	}
	var ads []engage.AdCreative
	pr, err := utils.PaginateWithCustomQuery(db, db.Order("display_order asc"), req.Page, req.PageSize, &ads)
	if err != nil {
		return nil, 0, 0, err
	}
	out := make([]engageResp.AdDetail, 0, len(ads))
	for _, ad := range ads {
		ctr := float64(0)
		if ad.Impressions > 0 {
			ctr = float64(ad.Clicks) / float64(ad.Impressions) * 100
		}
		out = append(out, engageResp.AdDetail{AdCreative: ad, CTR: ctr})
	}
	return out, pr.Total, pr.TotalPage, nil
}

// ActiveAds returns a paginated list of currently-active creatives for a placement.
func (a AdServices) ActiveAds(placement string, page, pageSize int) ([]engage.AdCreative, int64, int64, error) {
	if global.GVA_DB == nil {
		return nil, 0, 0, errors.New("db not init")
	}
	var list []engage.AdCreative
	countDB := global.GVA_DB.Model(&engage.AdCreative{}).
		Where("placement = ? AND status = ?", placement, constants.AdStatusActive)
	queryDB := global.GVA_DB.Model(&engage.AdCreative{}).
		Where("placement = ? AND status = ?", placement, constants.AdStatusActive).
		Order("display_order asc")
	pr, err := utils.PaginateWithCustomQuery(countDB, queryDB, page, pageSize, &list)
	if err != nil {
		return nil, 0, 0, err
	}
	return list, pr.Total, pr.TotalPage, nil
}

// GetAd returns a single ad by id.
func (a AdServices) GetAd(id uint) (*engage.AdCreative, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	var ad engage.AdCreative
	if err := global.GVA_DB.First(&ad, id).Error; err != nil {
		return nil, errors.New("ad not found")
	}
	return &ad, nil
}

// RunScheduler flips ads per their schedule: scheduled -> active at start, active -> expired at end.
// Called by the timer each tick.
func (a AdServices) RunScheduler() error {
	if global.GVA_DB == nil {
		return errors.New("db not init")
	}
	now := time.Now()
	// activate scheduled ads whose start has arrived
	_ = global.GVA_DB.Model(&engage.AdCreative{}).
		Where("status = ? AND start_datetime <= ? AND (end_datetime IS NULL OR end_datetime > ?)",
			constants.AdStatusScheduled, now, now).
		Update("status", constants.AdStatusActive).Error
	// expire active ads whose end has passed
	return global.GVA_DB.Model(&engage.AdCreative{}).
		Where("status = ? AND end_datetime IS NOT NULL AND end_datetime <= ?", constants.AdStatusActive, now).
		Update("status", constants.AdStatusExpired).Error
}

// validatePlacement enforces the max-active cap and a unique display order for a placement.
func (a AdServices) validatePlacement(req engageReq.AdCreativeReq) error {
	var activeCount int64
	_ = global.GVA_DB.Model(&engage.AdCreative{}).
		Where("placement = ? AND status IN ?", req.Placement, []string{constants.AdStatusActive, constants.AdStatusScheduled}).
		Count(&activeCount).Error
	if activeCount >= maxAdsPerPlacement {
		return errors.New("placement already has the maximum active/scheduled ads")
	}
	var dup int64
	_ = global.GVA_DB.Model(&engage.AdCreative{}).
		Where("placement = ? AND display_order = ?", req.Placement, req.DisplayOrder).Count(&dup).Error
	if dup > 0 {
		return errors.New("display order already used in this placement")
	}
	return nil
}
