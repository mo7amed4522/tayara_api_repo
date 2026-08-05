package engage

import (
	"github.com/gin-gonic/gin"
	commonReq "github.com/mo7amed4522/tayara_api/model/common/request"
	"github.com/mo7amed4522/tayara_api/model/common/response"
	engageReq "github.com/mo7amed4522/tayara_api/model/engage/request"
	"github.com/mo7amed4522/tayara_api/utils"
)

type AdApi struct{}

// CreateAd
// @Tags Engage
// @Summary Admin creates an ad creative (draft)
// @Produce application/json
// @Param data body engageReq.AdCreativeReq true "ad payload"
// @Success 200 {object} response.Response
// @Router /api/v1/engage/ads [post]
func (AdApi) CreateAd(c *gin.Context) {
	adminID := utils.GetUserID(c)
	var req engageReq.AdCreativeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	ad, err := engageService.AdServices.CreateAd(adminID, req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(ad, c)
}

// UpdateAd
// @Tags Engage
// @Summary Admin updates an ad creative
// @Produce application/json
// @Param data body engageReq.AdCreativeReq true "ad payload"
// @Success 200 {object} response.Response
// @Router /api/v1/engage/ads [put]
func (AdApi) UpdateAd(c *gin.Context) {
	adminID := utils.GetUserID(c)
	var req engageReq.AdCreativeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	ad, err := engageService.AdServices.UpdateAd(adminID, req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(ad, c)
}

// SetStatus
// @Tags Engage
// @Summary Admin activates/disables an ad
// @Produce application/json
// @Param data body engageReq.AdSetStatusReq true "status action"
// @Success 200 {object} response.Response
// @Router /api/v1/engage/ads/status [put]
func (AdApi) SetStatus(c *gin.Context) {
	var req engageReq.AdSetStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := engageService.AdServices.SetStatus(req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("ad status updated", c)
}

// DeleteAd
// @Tags Engage
// @Summary Admin deletes an ad (draft/disabled only)
// @Produce application/json
// @Param query query commonReq.GetById true "ad id"
// @Success 200 {object} response.Response
// @Router /api/v1/engage/ads [delete]
func (AdApi) DeleteAd(c *gin.Context) {
	var idReq commonReq.GetById
	if err := c.ShouldBindQuery(&idReq); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := engageService.AdServices.DeleteAd(idReq.Uint()); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("ad deleted", c)
}

// ListAds
// @Tags Engage
// @Summary Admin lists ads with CTR (filterable by placement/status)
// @Produce application/json
// @Param data query engageReq.AdCreativePageReq true "filter"
// @Success 200 {object} response.Response
// @Router /api/v1/engage/admin/ads [get]
func (AdApi) ListAds(c *gin.Context) {
	var req engageReq.AdCreativePageReq
	_ = c.ShouldBindQuery(&req)
	list, total, err := engageService.AdServices.ListAds(req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(response.PageResult{List: list, Total: total, Page: req.Page, PageSize: req.PageSize}, "ok", c)
}

// ActiveAds
// @Tags Engage
// @Summary Public list of currently-active ads for a placement
// @Produce application/json
// @Param placement query string true "placement"
// @Success 200 {object} response.Response
// @Router /api/v1/engage/public/ads [get]
func (AdApi) ActiveAds(c *gin.Context) {
	placement := c.Query("placement")
	if placement == "" {
		response.FailWithMessage("placement required", c)
		return
	}
	list, err := engageService.AdServices.ActiveAds(placement)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(list, c)
}

// RecordImpression
// @Tags Engage
// @Summary Public records an ad impression
// @Produce application/json
// @Param data body commonReq.GetById true "ad id"
// @Success 200 {object} response.Response
// @Router /api/v1/engage/public/ads/impression [post]
func (AdApi) RecordImpression(c *gin.Context) {
	var req commonReq.GetById
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := engageService.AdServices.RecordImpression(uint(req.ID)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("ok", c)
}

// RecordClick
// @Tags Engage
// @Summary Public records an ad click
// @Produce application/json
// @Param data body engageReq.AdClickReq true "click payload"
// @Success 200 {object} response.Response
// @Router /api/v1/engage/public/ads/click [post]
func (AdApi) RecordClick(c *gin.Context) {
	var req engageReq.AdClickReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	var userID *uint
	if id := utils.GetBxUserID(c); id > 0 {
		uid := id
		userID = &uid
	}
	click, err := engageService.AdServices.RecordClick(userID, req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(click, c)
}
