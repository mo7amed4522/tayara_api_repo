package catalog

import (
	"errors"
	"time"

	"github.com/mo7amed4522/tayara_api/constants"
	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/model/catalog"
	"github.com/mo7amed4522/tayara_api/model/catalog/request"
	"github.com/mo7amed4522/tayara_api/model/catalog/response"
	"github.com/mo7amed4522/tayara_api/model/dxapi"
	"github.com/mo7amed4522/tayara_api/model/engage"
	"github.com/mo7amed4522/tayara_api/utils"
	"gorm.io/gorm"
)

type PackageServices struct{}

// CreatePackage creates a package (status draft) owned by the authenticated agent, together with
// its child items, in one transaction. Discount is enforced to be at least the locked minimum.
func (p *PackageServices) CreatePackage(agentID uint, req request.PackageReq) (*catalog.Package, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	if err := p.validate(req); err != nil {
		return nil, err
	}
	if err := global.GVA_DB.First(&dxapi.DxAgentProfile{}, agentID).Error; err != nil {
		return nil, errors.New("agent not found")
	}
	pkg := &catalog.Package{
		PackageCode:          utils.NextCode(constants.CodePrefixPackage),
		AgentID:              agentID,
		Type:                 req.Type,
		Name:                 req.Name,
		CoverImageURL:        req.CoverImageURL,
		Description:          req.Description,
		GalleryImages:        req.GalleryImages,
		DestinationCountryID: req.DestinationCountryID,
		Location:             req.Location,
		Category:             req.Category,
		DaysDuration:         req.DaysDuration,
		AvailabilityType:     req.AvailabilityType,
		DisplayPrice:         req.DisplayPrice,
		EstimatedPrice:       req.EstimatedPrice,
		DiscountPercent:      req.DiscountPercent,
		Status:               constants.PackageStatusDraft,
		StatusTour:           req.StatusTour,
		TimeDisplay:          req.TimeDisplay,
		Destination:          req.Destination,
		OriginalPrice:        req.OriginalPrice,
		PromoTag:             req.PromoTag,
	}
	err := global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(pkg).Error; err != nil {
			return err
		}
		return insertChildren(tx, pkg.ID, req)
	})
	return pkg, err
}

// SubmitForApproval moves one of the agent's own packages to pending_approval.
func (p *PackageServices) SubmitForApproval(agentID, id uint) error {
	if global.GVA_DB == nil {
		return errors.New("db not init")
	}
	var pkg catalog.Package
	if err := global.GVA_DB.Where("id = ? AND agent_id = ?", id, agentID).First(&pkg).Error; err != nil {
		return errors.New("package not found")
	}
	if pkg.Status != constants.PackageStatusDraft {
		return errors.New("only draft packages can be submitted")
	}
	now := time.Now()
	return global.GVA_DB.Model(&catalog.Package{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":       constants.PackageStatusPendingApproval,
			"submitted_at": &now,
		}).Error
}

// UpdatePackage updates one of the agent's own packages (header + replaces children) in a transaction.
func (p *PackageServices) UpdatePackage(agentID uint, req request.PackageReq) (*response.PackageDetail, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	if err := p.validate(req); err != nil {
		return nil, err
	}
	var pkg catalog.Package
	if err := global.GVA_DB.Where("id = ? AND agent_id = ?", req.ID, agentID).First(&pkg).Error; err != nil {
		return nil, errors.New("package not found")
	}
	now := time.Now()
	updates := map[string]interface{}{
		"type":                   req.Type,
		"name":                   req.Name,
		"cover_image_url":        req.CoverImageURL,
		"description":            req.Description,
		"gallery_images":         req.GalleryImages,
		"destination_country_id": req.DestinationCountryID,
		"location":               req.Location,
		"category":               req.Category,
		"days_duration":          req.DaysDuration,
		"availability_type":      req.AvailabilityType,
		"display_price":          req.DisplayPrice,
		"estimated_price":        req.EstimatedPrice,
		"discount_percent":       req.DiscountPercent,
		"status_tour":            req.StatusTour,
		"time_display":           req.TimeDisplay,
		"destination":            req.Destination,
		"original_price":         req.OriginalPrice,
		"promo_tag":              req.PromoTag,
		"updated_at":             now,
	}
	err := global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&catalog.Package{}).Where("id = ?", req.ID).Updates(updates).Error; err != nil {
			return err
		}
		if err := clearChildren(tx, req.ID); err != nil {
			return err
		}
		return insertChildren(tx, req.ID, req)
	})
	if err != nil {
		return nil, err
	}
	return p.GetOwnPackage(agentID, req.ID)
}

// DeletePackage soft-deletes one of the agent's own packages.
func (p *PackageServices) DeletePackage(agentID, id uint) error {
	if global.GVA_DB == nil {
		return errors.New("db not init")
	}
	return global.GVA_DB.Where("id = ? AND agent_id = ?", id, agentID).Delete(&catalog.Package{}).Error
}

// GetOwnPackage returns one of the agent's own packages with children (agent scoping enforced).
func (p *PackageServices) GetOwnPackage(agentID, id uint) (*response.PackageDetail, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	return p.buildDetail(global.GVA_DB.Where("agent_id = ?", agentID), id)
}

// ListOwnPackages paginated list of the agent's own packages (agent scoping enforced).
func (p *PackageServices) ListOwnPackages(agentID uint, req request.PackagePageReq) ([]catalog.Package, int64, int64, error) {
	if global.GVA_DB == nil {
		return nil, 0, 0, errors.New("db not init")
	}
	db := global.GVA_DB.Model(&catalog.Package{}).Where("agent_id = ?", agentID)
	return p.page(db, req)
}

// ListPublicPackages paginated list of active packages for anonymous/buyer browsing.
func (p *PackageServices) ListPublicPackages(req request.PackagePageReq) ([]catalog.Package, int64, int64, error) {
	if global.GVA_DB == nil {
		return nil, 0, 0, errors.New("db not init")
	}
	db := global.GVA_DB.Model(&catalog.Package{}).Where("status = ?", constants.PackageStatusActive)
	if req.Type != "" {
		db = db.Where("type = ?", req.Type)
	}
	return p.page(db, req)
}

// GetPublicPackage returns an active package for buyers.
func (p *PackageServices) GetPublicPackage(id uint) (*response.PackageDetail, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	return p.buildDetail(global.GVA_DB.Where("status = ?", constants.PackageStatusActive), id)
}

// AdminListPackages paginated list of all packages (admin sees all).
func (p *PackageServices) AdminListPackages(req request.PackagePageReq) ([]catalog.Package, int64, int64, error) {
	if global.GVA_DB == nil {
		return nil, 0, 0, errors.New("db not init")
	}
	db := global.GVA_DB.Model(&catalog.Package{})
	return p.page(db, req)
}

// AdminGetPackageDetail returns a full package detail (header + children + reviews) for admin review.
func (p *PackageServices) AdminGetPackageDetail(id uint) (*response.PackageDetail, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	return p.buildDetail(global.GVA_DB, id)
}

// AdminSetStatus applies an admin moderation action to a package (active|hidden|flagged|disabled|pending_approval).
func (p *PackageServices) AdminSetStatus(adminID uint, req request.PackageAdminStatusReq) error {
	if global.GVA_DB == nil {
		return errors.New("db not init")
	}
	var pkg catalog.Package
	if err := global.GVA_DB.First(&pkg, req.ID).Error; err != nil {
		return errors.New("package not found")
	}
	now := time.Now()
	updates := map[string]interface{}{"status": req.Status, "admin_note": req.Note}
	if req.Status == constants.PackageStatusActive {
		updates["approved_by"] = adminID
		updates["approved_at"] = &now
	}
	if req.Status == constants.PackageStatusFlagged {
		updates["flag_issue_type"] = req.FlagIssueType
		updates["flag_details"] = req.FlagDetails
	}
	return global.GVA_DB.Model(&catalog.Package{}).Where("id = ?", req.ID).Updates(updates).Error
}

// validate enforces money/discount rules before persisting.
func (p *PackageServices) validate(req request.PackageReq) error {
	if req.Type != constants.PackageTypeInbound && req.Type != constants.PackageTypeOutbound {
		return errors.New("package type must be inbound or outbound")
	}
	if req.DiscountPercent < constants.DiscountMinPercentBps {
		return errors.New("discount must be at least 3%")
	}
	return nil
}

// buildDetail loads a package header + children under the given scoped query (already tied to caller).
func (p *PackageServices) buildDetail(scoped *gorm.DB, id uint) (*response.PackageDetail, error) {
	var pkg catalog.Package
	if err := scoped.First(&pkg, id).Error; err != nil {
		return nil, errors.New("package not found")
	}
	var (
		hotels       []catalog.PackageHotel
		activities   []catalog.PackageActivity
		addons       []catalog.PackageAddon
		avail        []catalog.PackageAvailability
		reqFields    []catalog.PackageRequiredField
		docReqs      []catalog.PackageDocumentRequirement
		itinDays     []catalog.ItineraryDay
		includes     []catalog.PackageInclusion
		excludes     []catalog.PackageExclusion
		highlights   []catalog.PackageHighlight
	)
	pid := pkg.ID
	if err := global.GVA_DB.Where("package_id = ?", pid).Find(&hotels).Error; err != nil {
		return nil, err
	}
	if err := global.GVA_DB.Where("package_id = ?", pid).Find(&activities).Error; err != nil {
		return nil, err
	}
	if err := global.GVA_DB.Where("package_id = ?", pid).Find(&addons).Error; err != nil {
		return nil, err
	}
	if err := global.GVA_DB.Where("package_id = ?", pid).Find(&avail).Error; err != nil {
		return nil, err
	}
	if err := global.GVA_DB.Where("package_id = ?", pid).Find(&reqFields).Error; err != nil {
		return nil, err
	}
	if err := global.GVA_DB.Where("package_id = ?", pid).Find(&docReqs).Error; err != nil {
		return nil, err
	}
	if err := global.GVA_DB.Where("package_id = ?", pid).Order("sort_order asc, day_number asc").Find(&itinDays).Error; err != nil {
		return nil, err
	}
	for i := range itinDays {
		var items []catalog.ItineraryItem
		if err := global.GVA_DB.Where("itinerary_day_id = ?", itinDays[i].ID).Order("sort_order asc").Find(&items).Error; err != nil {
			return nil, err
		}
		itinDays[i].Items = items
	}
	if err := global.GVA_DB.Where("package_id = ?", pid).Order("sort_order asc").Find(&includes).Error; err != nil {
		return nil, err
	}
	if err := global.GVA_DB.Where("package_id = ?", pid).Order("sort_order asc").Find(&excludes).Error; err != nil {
		return nil, err
	}
	if err := global.GVA_DB.Where("package_id = ?", pid).Order("sort_order asc").Find(&highlights).Error; err != nil {
		return nil, err
	}
	return &response.PackageDetail{
		Package:              pkg,
		Hotels:               hotels,
		Activities:           activities,
		Addons:               addons,
		Availability:         avail,
		RequiredFields:       reqFields,
		DocumentRequirements: docReqs,
		ItineraryDays:        itinDays,
		Includes:             includes,
		Excludes:             excludes,
		Highlights:           highlights,
	}, nil
}

// ListPackageReviews returns paginated reviews for a given package.
func (p *PackageServices) ListPackageReviews(packageID uint, page, pageSize int) ([]engage.Review, int64, int64, error) {
	if global.GVA_DB == nil {
		return nil, 0, 0, errors.New("db not init")
	}
	var rows []engage.Review
	var total int64
	if err := global.GVA_DB.Model(&engage.Review{}).Where("package_id = ?", packageID).Count(&total).Error; err != nil {
		return nil, 0, 0, err
	}
	db := global.GVA_DB.Where("package_id = ?", packageID).
		Order("created_at desc").
		Offset((page - 1) * pageSize).
		Limit(pageSize)
	if err := db.Find(&rows).Error; err != nil {
		return nil, 0, 0, err
	}
	return rows, total, int64((int(total)+pageSize-1)/pageSize), nil
}

// page runs a paginated count+find on the given model scope.
func (p *PackageServices) page(db *gorm.DB, req request.PackagePageReq) ([]catalog.Package, int64, int64, error) {
	var list []catalog.Package
	pr, err := utils.PaginateWithCustomQuery(db, db.Order("updated_at desc"), req.Page, req.PageSize, &list)
	if err != nil {
		return nil, 0, 0, err
	}
	return list, pr.Total, pr.TotalPage, nil
}

// clearChildren removes all child rows of a package (used on update before re-insert).
func clearChildren(tx *gorm.DB, pkgID uint) error {
	if err := tx.Where("package_id = ?", pkgID).Delete(&catalog.PackageHotel{}).Error; err != nil {
		return err
	}
	if err := tx.Where("package_id = ?", pkgID).Delete(&catalog.PackageActivity{}).Error; err != nil {
		return err
	}
	if err := tx.Where("package_id = ?", pkgID).Delete(&catalog.PackageAddon{}).Error; err != nil {
		return err
	}
	if err := tx.Where("package_id = ?", pkgID).Delete(&catalog.PackageAvailability{}).Error; err != nil {
		return err
	}
	if err := tx.Where("package_id = ?", pkgID).Delete(&catalog.PackageRequiredField{}).Error; err != nil {
		return err
	}
	if err := tx.Where("package_id = ?", pkgID).Delete(&catalog.PackageDocumentRequirement{}).Error; err != nil {
		return err
	}
	if err := tx.Where("package_id = ?", pkgID).Delete(&catalog.ItineraryDay{}).Error; err != nil {
		return err
	}
	if err := tx.Where("package_id = ?", pkgID).Delete(&catalog.PackageInclusion{}).Error; err != nil {
		return err
	}
	if err := tx.Where("package_id = ?", pkgID).Delete(&catalog.PackageExclusion{}).Error; err != nil {
		return err
	}
	if err := tx.Where("package_id = ?", pkgID).Delete(&catalog.PackageHighlight{}).Error; err != nil {
		return err
	}
	// Itinerary items are children of ItineraryDay, but we also clear them here for safety.
	if err := tx.Where("itinerary_day_id IN (SELECT id FROM catalog_itinerary_day WHERE package_id = ?)", pkgID).Delete(&catalog.ItineraryItem{}).Error; err != nil {
		return err
	}
	return nil
}

// insertChildren persists the child items of a package (called inside a transaction).
func insertChildren(tx *gorm.DB, pkgID uint, req request.PackageReq) error {
	if len(req.Hotels) > 0 {
		rows := make([]catalog.PackageHotel, 0, len(req.Hotels))
		for _, h := range req.Hotels {
			rows = append(rows, catalog.PackageHotel{PackageID: pkgID, HotelStars: h.HotelStars, RoomType: h.RoomType,
				OptionName: h.OptionName, PricePerNightEst: h.PricePerNightEst, IncludeInPackage: h.IncludeInPackage,
				Nights: h.Nights, SortOrder: h.SortOrder})
		}
		if err := tx.Create(&rows).Error; err != nil {
			return err
		}
	}
	if len(req.Activities) > 0 {
		rows := make([]catalog.PackageActivity, 0, len(req.Activities))
		for _, a := range req.Activities {
			rows = append(rows, catalog.PackageActivity{PackageID: pkgID, Title: a.Title, ImageURL: a.ImageURL,
				EstimatedCost: a.EstimatedCost, IncludeInPackage: a.IncludeInPackage, SortOrder: a.SortOrder})
		}
		if err := tx.Create(&rows).Error; err != nil {
			return err
		}
	}
	if len(req.Addons) > 0 {
		rows := make([]catalog.PackageAddon, 0, len(req.Addons))
		for _, a := range req.Addons {
			rows = append(rows, catalog.PackageAddon{PackageID: pkgID, AddonType: a.AddonType,
				EstimatedPrice: a.EstimatedPrice, PriceModel: a.PriceModel, IncludeInPackage: a.IncludeInPackage, Active: a.Active})
		}
		if err := tx.Create(&rows).Error; err != nil {
			return err
		}
	}
	if len(req.Availability) > 0 {
		rows := make([]catalog.PackageAvailability, 0, len(req.Availability))
		for _, a := range req.Availability {
			rows = append(rows, catalog.PackageAvailability{PackageID: pkgID, StartDate: parseDate(a.StartDate),
				EndDate: parseDate(a.EndDate), IsFixedGroup: a.IsFixedGroup, MaxSeats: a.MaxSeats, Notes: a.Notes})
		}
		if err := tx.Create(&rows).Error; err != nil {
			return err
		}
	}
	if len(req.RequiredFields) > 0 {
		rows := make([]catalog.PackageRequiredField, 0, len(req.RequiredFields))
		for _, f := range req.RequiredFields {
			rows = append(rows, catalog.PackageRequiredField{PackageID: pkgID, FieldCode: f.FieldCode,
				IsAgentSelectable: f.IsAgentSelectable, IsAdminLocked: f.IsAdminLocked, IsRequired: f.IsRequired, SortOrder: f.SortOrder})
		}
		if err := tx.Create(&rows).Error; err != nil {
			return err
		}
	}
	if len(req.DocumentRequirements) > 0 {
		rows := make([]catalog.PackageDocumentRequirement, 0, len(req.DocumentRequirements))
		for _, d := range req.DocumentRequirements {
			rows = append(rows, catalog.PackageDocumentRequirement{PackageID: pkgID, DocumentTypeID: d.DocumentTypeID,
				IsRequired: d.IsRequired, IsOptional: d.IsOptional, RequiredAfter: d.RequiredAfter, IsLocked: d.IsLocked, Notes: d.Notes})
		}
		if err := tx.Create(&rows).Error; err != nil {
			return err
		}
	}
	if len(req.ItineraryDays) > 0 {
		for _, day := range req.ItineraryDays {
			d := catalog.ItineraryDay{
				PackageID:   pkgID,
				DayNumber:   day.DayNumber,
				Title:       day.Title,
				Description: day.Description,
				SortOrder:   day.SortOrder,
			}
			if err := tx.Create(&d).Error; err != nil {
				return err
			}
			if len(day.Items) > 0 {
				items := make([]catalog.ItineraryItem, 0, len(day.Items))
				for _, it := range day.Items {
					items = append(items, catalog.ItineraryItem{
						ItineraryDayID: d.ID, Time: it.Time, Title: it.Title,
						Description: it.Description, SortOrder: it.SortOrder,
					})
				}
				if err := tx.Create(&items).Error; err != nil {
					return err
				}
			}
		}
	}
	if len(req.Includes) > 0 {
		rows := make([]catalog.PackageInclusion, 0, len(req.Includes))
		for _, inc := range req.Includes {
			rows = append(rows, catalog.PackageInclusion{PackageID: pkgID, Category: inc.Category,
				ItemText: inc.ItemText, SortOrder: inc.SortOrder})
		}
		if err := tx.Create(&rows).Error; err != nil {
			return err
		}
	}
	if len(req.Excludes) > 0 {
		rows := make([]catalog.PackageExclusion, 0, len(req.Excludes))
		for _, exc := range req.Excludes {
			rows = append(rows, catalog.PackageExclusion{PackageID: pkgID, ItemText: exc.ItemText,
				SortOrder: exc.SortOrder})
		}
		if err := tx.Create(&rows).Error; err != nil {
			return err
		}
	}
	if len(req.Highlights) > 0 {
		rows := make([]catalog.PackageHighlight, 0, len(req.Highlights))
		for _, h := range req.Highlights {
			rows = append(rows, catalog.PackageHighlight{PackageID: pkgID, Name: h.Name,
				IconURL: h.IconURL, SortOrder: h.SortOrder})
		}
		if err := tx.Create(&rows).Error; err != nil {
			return err
		}
	}
	return nil
}

// parseDate parses a 2006-01-02 string into a time pointer (nil if empty/malformed).
func parseDate(s string) *time.Time {
	if s == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return nil
	}
	return &t
}
