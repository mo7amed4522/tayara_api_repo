package api

import (
	"github.com/gin-gonic/gin"
	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/model/common/response"
	"github.com/mo7amed4522/tayara_api/plugin/announcement/model"
	"github.com/mo7amed4522/tayara_api/plugin/announcement/model/request"
	"go.uber.org/zap"
)

var Info = new(info)

type info struct{}

// CreateInfo Create announcement
// @Tags Info
// @Summary Create announcement
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body model.Info true "Create announcement"
// @Success 200 {object} response.Response{msg=string} "Create success"
// @Router /info/createInfo [post]
func (a *info) CreateInfo(c *gin.Context) {
	var info model.Info
	err := c.ShouldBindJSON(&info)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	err = serviceInfo.CreateInfo(&info)
	if err != nil {
		global.GVA_LOG.Error("Create failed!", zap.Error(err))
		response.FailWithMessage("Create failed", c)
		return
	}
	response.OkWithMessage("Create success", c)
}

// DeleteInfo Delete announcement
// @Tags Info
// @Summary Delete announcement
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body model.Info true "Delete announcement"
// @Success 200 {object} response.Response{msg=string} "Delete success"
// @Router /info/deleteInfo [delete]
func (a *info) DeleteInfo(c *gin.Context) {
	ID := c.Query("ID")
	err := serviceInfo.DeleteInfo(ID)
	if err != nil {
		global.GVA_LOG.Error("Delete failed!", zap.Error(err))
		response.FailWithMessage("Delete failed", c)
		return
	}
	response.OkWithMessage("Delete success", c)
}

// DeleteInfoByIds Batch delete announcement
// @Tags Info
// @Summary Batch delete announcement
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Success 200 {object} response.Response{msg=string} "Batch delete success"
// @Router /info/deleteInfoByIds [delete]
func (a *info) DeleteInfoByIds(c *gin.Context) {
	IDs := c.QueryArray("IDs[]")
	if err := serviceInfo.DeleteInfoByIds(IDs); err != nil {
		global.GVA_LOG.Error("Batch delete failed!", zap.Error(err))
		response.FailWithMessage("Batch delete failed", c)
		return
	}
	response.OkWithMessage("Batch delete success", c)
}

// UpdateInfo Update announcement
// @Tags Info
// @Summary Update announcement
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body model.Info true "Update announcement"
// @Success 200 {object} response.Response{msg=string} "Update success"
// @Router /info/updateInfo [put]
func (a *info) UpdateInfo(c *gin.Context) {
	var info model.Info
	err := c.ShouldBindJSON(&info)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	err = serviceInfo.UpdateInfo(info)
	if err != nil {
		global.GVA_LOG.Error("Update failed!", zap.Error(err))
		response.FailWithMessage("Update failed", c)
		return
	}
	response.OkWithMessage("Update success", c)
}

// FindInfo Use id to query announcement
// @Tags Info
// @Summary Use id to query announcement
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data query model.Info true "Use id to query announcement"
// @Success 200 {object} response.Response{data=model.Info,msg=string} "Find success"
// @Router /info/findInfo [get]
func (a *info) FindInfo(c *gin.Context) {
	ID := c.Query("ID")
	reinfo, err := serviceInfo.GetInfo(ID)
	if err != nil {
		global.GVA_LOG.Error("Find failed!", zap.Error(err))
		response.FailWithMessage("Find failed", c)
		return
	}
	response.OkWithData(reinfo, c)
}

// GetInfoList Get announcement list by page
// @Tags Info
// @Summary Get announcement list by page
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data query request.InfoSearch true "Get announcement list by page"
// @Success 200 {object} response.Response{data=response.PageResult,msg=string} "Get success"
// @Router /info/getInfoList [get]
func (a *info) GetInfoList(c *gin.Context) {
	var pageInfo request.InfoSearch
	err := c.ShouldBindQuery(&pageInfo)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := serviceInfo.GetInfoInfoList(pageInfo)
	if err != nil {
		global.GVA_LOG.Error("Get failed!", zap.Error(err))
		response.FailWithMessage("Get failed", c)
		return
	}
	response.OkWithDetailed(response.PageResult{
		List:     list,
		Total:    total,
		Page:     pageInfo.Page,
		PageSize: pageInfo.PageSize,
	}, "Get success", c)
}

// GetInfoDataSource Get Info data source
// @Tags Info
// @Summary Get Info data source
// @accept application/json
// @Produce application/json
// @Success 200 {object} response.Response{data=object,msg=string} "Find success"
// @Router /info/getInfoDataSource [get]
func (a *info) GetInfoDataSource(c *gin.Context) {
	// This interface is to define the data source
	dataSource, err := serviceInfo.GetInfoDataSource()
	if err != nil {
		global.GVA_LOG.Error("Find failed!", zap.Error(err))
		response.FailWithMessage("Find failed", c)
		return
	}
	response.OkWithData(dataSource, c)
}

// GetInfoPublic Announcement interface without authentication
// @Tags Info
// @Summary Announcement interface without authentication
// @accept application/json
// @Produce application/json
// @Param data query request.InfoSearch true "Get announcement list by page"
// @Success 200 {object} response.Response{data=object,msg=string} "Get success"
// @Router /info/getInfoPublic [get]
func (a *info) GetInfoPublic(c *gin.Context) {
	// This interface does not need authentication The example is a fixed message interface, generally this interface is used for C-end services, and the business logic needs to be implemented by yourself
	response.OkWithDetailed(gin.H{"info": "Announcement interface without authentication information"}, "Get success", c)
}
