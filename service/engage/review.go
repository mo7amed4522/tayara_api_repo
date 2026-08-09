package engage

import (
	"errors"

	"github.com/mo7amed4522/tayara_api/constants"
	"github.com/mo7amed4522/tayara_api/global"
	bookingModel "github.com/mo7amed4522/tayara_api/model/booking"
	"github.com/mo7amed4522/tayara_api/model/engage"
	engageReq "github.com/mo7amed4522/tayara_api/model/engage/request"
	"github.com/mo7amed4522/tayara_api/utils"
)

type ReviewServices struct{}

// CreateReview records a post-completion review & rating. Only an active (completed) booking's
// owner may review. Returns an error if the booking is not completed.
func (r ReviewServices) CreateReview(userID uint, req engageReq.ReviewReq) (*engage.Review, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	var bk bookingModel.Booking
	if err := global.GVA_DB.Where("id = ? AND user_id = ?", req.BookingID, userID).First(&bk).Error; err != nil {
		return nil, errors.New("booking not found")
	}
	if bk.BookingStatus != constants.BookingStatusCompleted {
		return nil, errors.New("reviews are only allowed on completed bookings")
	}
	var agentID, pkgID *uint
	if bk.AgentID > 0 {
		id := bk.AgentID
		agentID = &id
	}
	if bk.PackageID != nil {
		id := *bk.PackageID
		pkgID = &id
	}
	review := &engage.Review{
		BookingID: req.BookingID, UserID: userID, AgentID: agentID, PackageID: pkgID,
		Rating: req.Rating, ReviewText: req.ReviewText, ReviewStatus: constants.ReviewStatusActive,
	}
	if err := global.GVA_DB.Create(review).Error; err != nil {
		return nil, err
	}
	return review, nil
}

// ListReviews returns reviews filtered by agent/package/status (admin sees all states).
func (r ReviewServices) ListReviews(req engageReq.ReviewPageReq) ([]engage.Review, int64, int64, error) {
	if global.GVA_DB == nil {
		return nil, 0, 0, errors.New("db not init")
	}
	db := global.GVA_DB.Model(&engage.Review{})
	if req.AgentID > 0 {
		db = db.Where("agent_id = ?", req.AgentID)
	}
	if req.PackageID > 0 {
		db = db.Where("package_id = ?", req.PackageID)
	}
	if req.Status != "" {
		db = db.Where("review_status = ?", req.Status)
	}
	var list []engage.Review
	pr, err := utils.PaginateWithCustomQuery(db, db.Order("created_at desc"), req.Page, req.PageSize, &list)
	if err != nil {
		return nil, 0, 0, err
	}
	return list, pr.Total, pr.TotalPage, nil
}

// PublicReviews returns only active published reviews (buyer-facing).
func (r ReviewServices) PublicReviews(req engageReq.ReviewPageReq) ([]engage.Review, int64, int64, error) {
	req.Status = constants.ReviewStatusActive
	return r.ListReviews(req)
}

// SetStatus lets admin moderate a review (disable/active).
func (r ReviewServices) SetStatus(adminID uint, reviewID uint, status string) error {
	if global.GVA_DB == nil {
		return errors.New("db not init")
	}
	return global.GVA_DB.Model(&engage.Review{}).Where("id = ?", reviewID).
		Update("review_status", status).Error
}
