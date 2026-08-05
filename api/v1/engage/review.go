package engage

import (
	"github.com/gin-gonic/gin"
	commonReq "github.com/mo7amed4522/tayara_api/model/common/request"
	"github.com/mo7amed4522/tayara_api/model/common/response"
	engageReq "github.com/mo7amed4522/tayara_api/model/engage/request"
	"github.com/mo7amed4522/tayara_api/utils"
)

type ReviewApi struct{}

// CreateReview
// @Tags Engage
// @Summary User reviews a completed booking (rating + text)
// @Produce application/json
// @Param data body engageReq.ReviewReq true "review payload"
// @Success 200 {object} response.Response
// @Router /api/v1/engage/reviews [post]
func (ReviewApi) CreateReview(c *gin.Context) {
	userID := utils.GetBxUserID(c)
	var req engageReq.ReviewReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	review, err := engageService.ReviewServices.CreateReview(userID, req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(review, c)
}

// ListReviews
// @Tags Review
// @Summary Public list of active reviews (filterable)
// @Produce application/json
// @Param data query engageReq.ReviewPageReq true "filter"
// @Success 200 {object} response.Response
// @Router /api/v1/engage/public/reviews [get]
func (ReviewApi) ListReviews(c *gin.Context) {
	var req engageReq.ReviewPageReq
	_ = c.ShouldBindQuery(&req)
	list, total, err := engageService.ReviewServices.PublicReviews(req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(response.PageResult{List: list, Total: total, Page: req.Page, PageSize: req.PageSize}, "ok", c)
}

// AdminListReviews
// @Tags Review
// @Summary Admin lists all reviews (incl. disabled) for moderation
// @Produce application/json
// @Param data query engageReq.ReviewPageReq true "filter"
// @Success 200 {object} response.Response
// @Router /api/v1/engage/admin/reviews [get]
func (ReviewApi) AdminListReviews(c *gin.Context) {
	var req engageReq.ReviewPageReq
	_ = c.ShouldBindQuery(&req)
	list, total, err := engageService.ReviewServices.ListReviews(req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(response.PageResult{List: list, Total: total, Page: req.Page, PageSize: req.PageSize}, "ok", c)
}

// SetStatus
// @Tags Review
// @Summary Admin moderates a review status (active/disabled)
// @Produce application/json
// @Param query query commonReq.GetById true "review id"
// @Param status query string true "target status"
// @Success 200 {object} response.Response
// @Router /api/v1/engage/admin/reviews/status [put]
func (ReviewApi) SetStatus(c *gin.Context) {
	adminID := utils.GetUserID(c)
	var idReq commonReq.GetById
	if err := c.ShouldBindQuery(&idReq); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	status := c.Query("status")
	if status == "" {
		response.FailWithMessage("status required", c)
		return
	}
	if err := engageService.ReviewServices.SetStatus(adminID, idReq.Uint(), status); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("review status updated", c)
}
