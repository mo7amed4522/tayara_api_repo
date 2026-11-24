package bxapi

import (
	"fmt"

	"github.com/gin-gonic/gin"
	bxapiReq "github.com/mo7amed4522/tayara_api/model/bxapi/request"
	bxapiResp "github.com/mo7amed4522/tayara_api/model/bxapi/response"
	response "github.com/mo7amed4522/tayara_api/model/common/response"
	"github.com/mo7amed4522/tayara_api/utils"
)

type HomeApi struct{}

// CustomerSearch
//
//	@Tags		Customer
//	@Summary	Public search across agents and packages
//	@Produce	application/json
//	@Param		keyword		query		string	true	"search keyword"
//	@Param		page		query		int		false	"page number"	default(1)
//	@Param		pageSize	query		int		false	"page size"		default(10)
//	@Success	200			{object}	response.Response
//	@Router		/api/v1/customer/search [get]
func (HomeApi) CustomerSearch(c *gin.Context) {
	var req bxapiReq.CustomerSearchReq
	if err := c.ShouldBindQuery(&req); err != nil || req.Keyword == "" {
		response.FailWithMessage("keyword is required", c)
		return
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	res, err := bxapiSvc.HomeServices.CustomerSearch(req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(res, c)
}

// ListExploreTrips
//
//	@Tags		Customer
//	@Summary	Public explore trips listing
//	@Produce	application/json
//	@Param		category	query		string	false	"category filter"
//	@Param		page		query		int		false	"page number"	default(1)
//	@Param		pageSize	query		int		false	"page size"		default(12)
//	@Success	200			{object}	response.Response
//	@Router		/api/v1/customer/explore/trips [get]
func (HomeApi) ListExploreTrips(c *gin.Context) {
	var req bxapiReq.CustomerHomeExploreReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 12
	}
	result, err := bxapiSvc.HomeServices.ListExploreTrips(req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(response.PageResult{List: result.Items, Total: result.Total, TotalPage: result.TotalPage, Page: req.Page, PageSize: req.PageSize}, "ok", c)
}

// ListDestinations
//
//	@Tags		Customer
//	@Summary	Public discover destinations
//	@Produce	application/json
//	@Param		page		query		int	false	"page number"	default(1)
//	@Param		pageSize	query		int	false	"page size"		default(10)
//	@Success	200			{object}	response.Response
//	@Router		/api/v1/customer/home/destinations [get]
func (HomeApi) ListDestinations(c *gin.Context) {
	var req bxapiReq.CustomerHomeDestinationsReq
	_ = c.ShouldBindQuery(&req)
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	paginateResult, err := bxapiSvc.HomeServices.ListDestinations(req.Page, req.PageSize)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	// Extract list from paginateResult.Data
	list, ok := paginateResult.Data.([]bxapiResp.CustomerDestinationCard)
	if !ok {
		response.FailWithMessage("unexpected destination pagination result", c)
		return
	}
	// Build page result with calculated total pages
	response.OkWithDetailed(response.PageResult{
		List:      list,
		Total:     paginateResult.Total,
		TotalPage: paginateResult.TotalPage,
		Page:      paginateResult.Page,
		PageSize:  paginateResult.PageSize,
	}, "ok", c)
}

// ListDeals
//
//	@Tags		Customer
//	@Summary	Public deals listing (discounted packages)
//	@Produce	application/json
//	@Param		category	query		string	false	"category filter"
//	@Param		page		query		int		false	"page number"	default(1)
//	@Param		pageSize	query		int		false	"page size"		default(12)
//	@Success	200			{object}	response.Response
//	@Router		/api/v1/customer/home/deals [get]
func (HomeApi) ListDeals(c *gin.Context) {
	var req bxapiReq.CustomerHomeDealsReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 12
	}
	result, err := bxapiSvc.HomeServices.ListDeals(req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(response.PageResult{List: result.Items, Total: result.Total, TotalPage: result.TotalPage, Page: req.Page, PageSize: req.PageSize}, "ok", c)
}

// ListTopRatedAgents
//
//	@Tags		Customer
//	@Summary	Public top rated agencies
//	@Produce	application/json
//	@Param		page		query		int	false	"page number"					default(1)
//	@Param		pageSize	query		int	false	"page size"						default(10)
//	@Param		limit		query		int	false	"deprecated page-size alias"	default(10)
//	@Success	200			{object}	response.Response
//	@Router		/api/v1/customer/home/agents [get]
func (HomeApi) ListTopRatedAgents(c *gin.Context) {
	page := utils.StringToInt(c.Query("page"))
	pageSize := utils.StringToInt(c.Query("pageSize"))
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = utils.StringToInt(c.Query("limit"))
	}
	if pageSize <= 0 || pageSize > 50 {
		pageSize = 10
	}
	result, err := bxapiSvc.HomeServices.ListTopRatedAgents(page, pageSize)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, ok := result.Data.([]bxapiResp.CustomerAgentCard)
	if !ok {
		response.FailWithMessage("unexpected agent pagination result", c)
		return
	}
	response.OkWithDetailed(response.PageResult{
		List: list, Total: result.Total, TotalPage: result.TotalPage,
		Page: result.Page, PageSize: result.PageSize,
	}, "ok", c)
}

// RecommendedForYou
//
//	@Tags		Customer
//	@Summary	Recommended packages for the authenticated user (wishlist + top picks)
//	@Produce	application/json
//	@Param		limit	query		int	false	"max results"	default(10)
//	@Success	200		{object}	response.Response
//	@Router		/api/v1/customer/home/recommended [get]
func (HomeApi) RecommendedForYou(c *gin.Context) {
	limit := 10
	if l := c.Query("limit"); l != "" {
		var n int
		if _, err := fmt.Sscanf(l, "%d", &n); err == nil && n > 0 && n <= 50 {
			limit = n
		}
	}
	userID := utils.GetBxUserID(c)
	list, err := bxapiSvc.HomeServices.RecommendedForYou(userID, limit)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(list, c)
}

// ActiveHomeAds
//
//	@Tags		Customer
//	@Summary	Public home top-banner ads
//	@Produce	application/json
//	@Param		page		query		int	false	"page number"	default(1)
//	@Param		pageSize	query		int	false	"page size"		default(10)
//	@Success	200			{object}	response.Response
//	@Router		/api/v1/customer/home/ads [get]
func (HomeApi) ActiveHomeAds(c *gin.Context) {
	var req bxapiReq.CustomerHomeAdsReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	result, err := bxapiSvc.HomeServices.ActiveHomeAds(req.Page, req.PageSize)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(response.PageResult{List: result.Items, Total: result.Total, TotalPage: result.TotalPage, Page: req.Page, PageSize: req.PageSize}, "ok", c)
}
