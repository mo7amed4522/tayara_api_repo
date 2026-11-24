package bxapi

import (
	"context"
	"errors"

	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/model/catalog"
	"github.com/mo7amed4522/tayara_api/model/catalog/response"
	"github.com/mo7amed4522/tayara_api/model/engage"
	"github.com/mo7amed4522/tayara_api/utils"
	"github.com/mo7amed4522/tayara_api/utils/cache"
	"gorm.io/gorm"
)

// PackageDetailServices owns the buyer-facing package detail operations.
type PackageDetailServices struct{}

type PackageReviewsResult struct {
	Items     []engage.Review `json:"items"`
	Total     int64           `json:"total"`
	TotalPage int64           `json:"totalPage"`
}

// GetPackageDetail returns a full package detail (header + children + reviews) for the customer app.
func (s *PackageDetailServices) GetPackageDetail(id uint) (*response.PackageDetail, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	detail, _, err := cache.GetOrLoadJSON(context.Background(), PackageDetailKey(id), sharedResponseTTL, func() (*response.PackageDetail, error) {
		return packageDetailServices.GetPublicPackage(id)
	})
	return detail, err
}

// ListPackageReviews returns paginated reviews for a customer-facing package detail page.
func (s *PackageDetailServices) ListPackageReviews(packageID uint, page, pageSize int) (PackageReviewsResult, error) {
	if global.GVA_DB == nil {
		return PackageReviewsResult{}, errors.New("db not init")
	}
	result, _, err := cache.GetOrLoadJSON(context.Background(), PackageReviewsKey(packageID, page, pageSize), sharedResponseTTL, func() (PackageReviewsResult, error) {
		items, total, totalPage, err := packageDetailServices.ListPackageReviewsDirect(packageID, page, pageSize)
		if err != nil {
			return PackageReviewsResult{}, err
		}
		return PackageReviewsResult{Items: items, Total: total, TotalPage: totalPage}, nil
	})
	return result, err
}

// Direct implementations that mirror the catalog service logic.

func (s *PackageDetailServices) GetPublicPackage(id uint) (*response.PackageDetail, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	return s.buildDetail(global.GVA_DB.Where("status = ?", "active"), id)
}

func (s *PackageDetailServices) ListPackageReviewsDirect(packageID uint, page, pageSize int) ([]engage.Review, int64, int64, error) {
	if global.GVA_DB == nil {
		return nil, 0, 0, errors.New("db not init")
	}
	var rows []engage.Review
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	countDB := global.GVA_DB.Model(&engage.Review{}).Where("package_id = ?", packageID)
	queryDB := global.GVA_DB.Model(&engage.Review{}).Where("package_id = ?", packageID).Order("created_at desc")

	paginateResult, err := utils.PaginateWithCustomQuery(countDB, queryDB, page, pageSize, &rows)
	if err != nil {
		return nil, 0, 0, err
	}
	total := paginateResult.Total
	totalPage := paginateResult.TotalPage
	return rows, total, int64(totalPage), nil
}

func (s *PackageDetailServices) buildDetail(scoped *gorm.DB, id uint) (*response.PackageDetail, error) {
	var pkg catalog.Package
	if err := scoped.First(&pkg, id).Error; err != nil {
		return nil, errors.New("package not found")
	}
	var (
		hotels     []catalog.PackageHotel
		activities []catalog.PackageActivity
		addons     []catalog.PackageAddon
		avail      []catalog.PackageAvailability
		reqFields  []catalog.PackageRequiredField
		docReqs    []catalog.PackageDocumentRequirement
		itinDays   []catalog.ItineraryDay
		includes   []catalog.PackageInclusion
		excludes   []catalog.PackageExclusion
		highlights []catalog.PackageHighlight
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

var packageDetailServices = &PackageDetailServices{}
