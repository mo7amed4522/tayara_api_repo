package catalog

import (
	"github.com/gin-gonic/gin"
	catalogReq "github.com/mo7amed4522/tayara_api/model/catalog/request"
	commonReq "github.com/mo7amed4522/tayara_api/model/common/request"
	"github.com/mo7amed4522/tayara_api/model/common/response"
	"github.com/mo7amed4522/tayara_api/utils"
)

type CatalogApi struct{}

// CreatePackage
//	@Tags		Catalog
//	@Summary	Agent creates a package (draft)
//	@Produce	application/json
//	@Param		data	body		catalogReq.PackageReq	true	"package payload"
//	@Success	200		{object}	response.Response
//	@Router		/api/v1/catalog/packages [post]
func (CatalogApi) CreatePackage(c *gin.Context) {
	agentID := utils.GetAgentID(c)
	if agentID == 0 {
		response.FailWithMessage("unauthorized agent", c)
		return
	}
	var req catalogReq.PackageReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	pkg, err := catalogService.PackageServices.CreatePackage(agentID, req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(pkg, c)
}

// UpdatePackage
//	@Tags		Catalog
//	@Summary	Agent updates one of their own packages
//	@Produce	application/json
//	@Param		data	body		catalogReq.PackageReq	true	"package payload"
//	@Success	200		{object}	response.Response
//	@Router		/api/v1/catalog/packages [put]
func (CatalogApi) UpdatePackage(c *gin.Context) {
	agentID := utils.GetAgentID(c)
	if agentID == 0 {
		response.FailWithMessage("unauthorized agent", c)
		return
	}
	var req catalogReq.PackageReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	pkg, err := catalogService.PackageServices.UpdatePackage(agentID, req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(pkg, c)
}

// SubmitForApproval
//	@Tags		Catalog
//	@Summary	Agent submits a draft package for approval
//	@Produce	application/json
//	@Param		id	query		int	true	"package id"
//	@Success	200	{object}	response.Response
//	@Router		/api/v1/catalog/packages/submit [post]
func (CatalogApi) SubmitForApproval(c *gin.Context) {
	agentID := utils.GetAgentID(c)
	var idReq commonReq.GetById
	if err := c.ShouldBindQuery(&idReq); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := catalogService.PackageServices.SubmitForApproval(agentID, uint(idReq.ID)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("submitted for approval", c)
}

// DeletePackage
//	@Tags		Catalog
//	@Summary	Agent deletes one of their own packages
//	@Produce	application/json
//	@Param		id	query		int	true	"package id"
//	@Success	200	{object}	response.Response
//	@Router		/api/v1/catalog/packages [delete]
func (CatalogApi) DeletePackage(c *gin.Context) {
	agentID := utils.GetAgentID(c)
	var idReq commonReq.GetById
	if err := c.ShouldBindQuery(&idReq); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := catalogService.PackageServices.DeletePackage(agentID, uint(idReq.ID)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("package deleted", c)
}

// GetOwnPackage
//	@Tags		Catalog
//	@Summary	Agent gets one of their own packages (with children)
//	@Produce	application/json
//	@Param		id	query		int	true	"package id"
//	@Success	200	{object}	response.Response
//	@Router		/api/v1/catalog/packages/detail [get]
func (CatalogApi) GetOwnPackage(c *gin.Context) {
	agentID := utils.GetAgentID(c)
	var idReq commonReq.GetById
	if err := c.ShouldBindQuery(&idReq); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	detail, err := catalogService.PackageServices.GetOwnPackage(agentID, uint(idReq.ID))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(detail, c)
}

// ListOwnPackages
//	@Tags		Catalog
//	@Summary	Agent lists their own packages (paginated)
//	@Produce	application/json
//	@Param		data	query		catalogReq.PackagePageReq	true	"page filter"
//	@Success	200		{object}	response.Response
//	@Router		/api/v1/catalog/packages [get]
func (CatalogApi) ListOwnPackages(c *gin.Context) {
	agentID := utils.GetAgentID(c)
	var req catalogReq.PackagePageReq
	_ = c.ShouldBindQuery(&req)
	list, total, totalPage, err := catalogService.PackageServices.ListOwnPackages(agentID, req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(response.PageResult{List: list, Total: total, TotalPage: totalPage, Page: req.Page, PageSize: req.PageSize}, "ok", c)
}

// ListPublicPackages
//	@Tags		Catalog
//	@Summary	Public list of active packages
//	@Produce	application/json
//	@Param		data	query		catalogReq.PackagePageReq	true	"page filter"
//	@Success	200		{object}	response.Response
//	@Router		/api/v1/catalog/public/packages [get]
func (CatalogApi) ListPublicPackages(c *gin.Context) {
	var req catalogReq.PackagePageReq
	_ = c.ShouldBindQuery(&req)
	list, total, totalPage, err := catalogService.PackageServices.ListPublicPackages(req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(response.PageResult{List: list, Total: total, TotalPage: totalPage, Page: req.Page, PageSize: req.PageSize}, "ok", c)
}

// GetPublicPackage
//	@Tags		Catalog
//	@Summary	Public detail of an active package
//	@Produce	application/json
//	@Param		id	query		int	true	"package id"
//	@Success	200	{object}	response.Response
//	@Router		/api/v1/catalog/public/packages/detail [get]
func (CatalogApi) GetPublicPackage(c *gin.Context) {
	var idReq commonReq.GetById
	if err := c.ShouldBindQuery(&idReq); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	detail, err := catalogService.PackageServices.GetPublicPackage(uint(idReq.ID))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(detail, c)
}

// ListPublicPackageReviews
//	@Tags		Catalog
//	@Summary	Public paginated reviews for an active package
//	@Produce	application/json
//	@Param		id			query		int	true	"package id"
//	@Param		page		query		int	false	"page number"	default(1)
//	@Param		pageSize	query		int	false	"page size"		default(10)
//	@Success	200			{object}	response.Response
//	@Router		/api/v1/catalog/public/packages/detail/reviews [get]
func (CatalogApi) ListPublicPackageReviews(c *gin.Context) {
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
	rows, total, totalPage, err := catalogService.PackageServices.ListPackageReviews(uint(idReq.ID), page, pageSize)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(response.PageResult{List: rows, Total: total, TotalPage: totalPage, Page: page, PageSize: pageSize}, "ok", c)
}

// AdminListPackages
//	@Tags		Catalog
//	@Summary	Admin lists all packages (moderation)
//	@Produce	application/json
//	@Param		data	query		catalogReq.PackagePageReq	true	"page filter"
//	@Success	200		{object}	response.Response
//	@Router		/api/v1/catalog/admin/packages [get]
func (CatalogApi) AdminListPackages(c *gin.Context) {
	var req catalogReq.PackagePageReq
	_ = c.ShouldBindQuery(&req)
	list, total, totalPage, err := catalogService.PackageServices.AdminListPackages(req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(response.PageResult{List: list, Total: total, TotalPage: totalPage, Page: req.Page, PageSize: req.PageSize}, "ok", c)
}

// AdminGetPackageDetail
//	@Tags		Catalog
//	@Summary	Admin gets full package detail (with children) for moderation review
//	@Produce	application/json
//	@Param		id	query		int	true	"package id"
//	@Success	200	{object}	response.Response
//	@Router		/api/v1/catalog/admin/packages/detail [get]
func (CatalogApi) AdminGetPackageDetail(c *gin.Context) {
	var idReq commonReq.GetById
	if err := c.ShouldBindQuery(&idReq); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	detail, err := catalogService.PackageServices.AdminGetPackageDetail(uint(idReq.ID))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(detail, c)
}

// AdminSetStatus
//	@Tags		Catalog
//	@Summary	Admin moderates a package (approve/hide/flag/disable)
//	@Produce	application/json
//	@Param		data	body		catalogReq.PackageAdminStatusReq	true	"status action"
//	@Success	200		{object}	response.Response
//	@Router		/api/v1/catalog/admin/packages/status [put]
func (CatalogApi) AdminSetStatus(c *gin.Context) {
	adminID := utils.GetUserID(c)
	var req catalogReq.PackageAdminStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := catalogService.PackageServices.AdminSetStatus(adminID, req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("package status updated", c)
}
