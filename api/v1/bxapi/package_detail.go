package bxapi

import (
	"github.com/gin-gonic/gin"
	commonReq "github.com/mo7amed4522/tayara_api/model/common/request"
	"github.com/mo7amed4522/tayara_api/model/common/response"
	"github.com/mo7amed4522/tayara_api/utils"
)

type PackageDetailApi struct{}

// GetPackageDetail
//
//	@Tags		Customer
//	@Summary	Public detail of an active package
//	@Produce	application/json
//	@Param		id	query		int	true	"package id"
//	@Success	200	{object}	response.Response
//	@Router		/api/v1/customer/packages/detail [get]
func (PackageDetailApi) GetPackageDetail(c *gin.Context) {
	var idReq commonReq.GetById
	if err := c.ShouldBindQuery(&idReq); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	detail, err := bxapiSvc.GetPackageDetail(uint(idReq.ID))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(detail, c)
}

// ListPackageReviews
//
//	@Tags		Customer
//	@Summary	Public paginated reviews for an active package
//	@Produce	application/json
//	@Param		id			query		int	true	"package id"
//	@Param		page		query		int	false	"page number"	default(1)
//	@Param		pageSize	query		int	false	"page size"		default(10)
//	@Success	200			{object}	response.Response
//	@Router		/api/v1/customer/packages/detail/reviews [get]
func (PackageDetailApi) ListPackageReviews(c *gin.Context) {
	var idReq commonReq.GetById
	if err := c.ShouldBindQuery(&idReq); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	page := utils.StringToInt(c.Query("page"))
	pageSize := utils.StringToInt(c.Query("pageSize"))
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	result, err := bxapiSvc.ListPackageReviews(uint(idReq.ID), page, pageSize)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(response.PageResult{List: result.Items, Total: result.Total, TotalPage: result.TotalPage, Page: page, PageSize: pageSize}, "ok", c)
}
