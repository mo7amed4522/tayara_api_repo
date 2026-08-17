package bxapi

import (
	"errors"

	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/model/catalog"
	"github.com/mo7amed4522/tayara_api/model/catalog/response"
	"github.com/mo7amed4522/tayara_api/model/engage"
	"gorm.io/gorm"
)

// PackageDetailServices owns the buyer-facing package detail operations.
type PackageDetailServices struct{}

// GetPackageDetail returns a full package detail (header + children + reviews) for the customer app.
func (s *PackageDetailServices) GetPackageDetail(id uint) (*response.PackageDetail, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	return packageDetailServices.GetPublicPackage(id)
}

// ListPackageReviews returns paginated reviews for a customer-facing package detail page.
func (s *PackageDetailServices) ListPackageReviews(packageID uint, page, pageSize int) ([]engage.Review, int64, int64, error) {
	if global.GVA_DB == nil {
		return nil, 0, 0, errors.New("db not init")
	}
	return packageDetailServices.ListPackageReviewsDirect(packageID, page, pageSize)
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

func (s *PackageDetailServices) buildDetail(scoped *gorm.DB, id uint) (*response.PackageDetail, error) {
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

var packageDetailServices = &PackageDetailServices{}
