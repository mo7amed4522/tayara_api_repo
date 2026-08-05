package dxapi

import (
	"github.com/gin-gonic/gin"
	commonReq "github.com/mo7amed4522/tayara_api/model/common/request"
	"github.com/mo7amed4522/tayara_api/model/common/response"
	dxapiReq "github.com/mo7amed4522/tayara_api/model/dxapi/request"
	"github.com/mo7amed4522/tayara_api/utils"
)

type AgentApi struct{}

// Register
// @Tags Agent
// @Summary Public agent (travel agency) registration
// @Produce application/json
// @Param data body dxapiReq.AgentRegisterReq true "agent payload"
// @Success 200 {object} response.Response
// @Router /api/v1/agent/register [post]
func (AgentApi) Register(c *gin.Context) {
	var req dxapiReq.AgentRegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	agent, err := agentService.AgentServices.Register(req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(agent, c)
}

// Login
// @Tags Agent
// @Summary Agent login (returns A-token)
// @Produce application/json
// @Param data body dxapiReq.AgentLoginReq true "login"
// @Success 200 {object} response.Response
// @Router /api/v1/agent/login [post]
func (AgentApi) Login(c *gin.Context) {
	var req dxapiReq.AgentLoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	res, err := agentService.AgentServices.Login(req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(res, c)
}

// GetProfile
// @Tags Agent
// @Summary Agent gets their own profile
// @Produce application/json
// @Success 200 {object} response.Response
// @Router /api/v1/agent/profile [get]
func (AgentApi) GetProfile(c *gin.Context) {
	agentID := utils.GetAgentID(c)
	agent, err := agentService.AgentServices.GetProfile(agentID)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(agent, c)
}

// ChangePassword
// @Tags Agent
// @Summary Agent changes their own password
// @Produce application/json
// @Param data body dxapiReq.ChangeAgentPasswordReq true "password change"
// @Success 200 {object} response.Response
// @Router /api/v1/agent/change-password [post]
func (AgentApi) ChangePassword(c *gin.Context) {
	agentID := utils.GetAgentID(c)
	var req dxapiReq.ChangeAgentPasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := agentService.AgentServices.ChangePassword(agentID, req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("password changed", c)
}

// ListAgents
// @Tags Agent
// @Summary Admin lists agents (paginated, filterable by status)
// @Produce application/json
// @Param data query dxapiReq.AgentPageReq true "page filter"
// @Success 200 {object} response.Response
// @Router /api/v1/agent/admin [get]
func (AgentApi) ListAgents(c *gin.Context) {
	var req dxapiReq.AgentPageReq
	_ = c.ShouldBindQuery(&req)
	list, total, err := agentService.AgentServices.List(req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(response.PageResult{List: list, Total: total, Page: req.Page, PageSize: req.PageSize}, "ok", c)
}

// GetAdminDetail
// @Tags Agent
// @Summary Admin gets an agent's full detail (documents + notes)
// @Produce application/json
// @Param id query int true "agent id"
// @Success 200 {object} response.Response
// @Router /api/v1/agent/admin/detail [get]
func (AgentApi) GetAdminDetail(c *gin.Context) {
	var idReq commonReq.GetById
	if err := c.ShouldBindQuery(&idReq); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	detail, err := agentService.AgentServices.GetAdminDetail(idReq.Uint())
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(detail, c)
}

// Approve
// @Tags Agent
// @Summary Admin approves a pending agent (sets commission % + SLA)
// @Produce application/json
// @Param data body dxapiReq.AgentApproveReq true "approve payload"
// @Success 200 {object} response.Response
// @Router /api/v1/agent/admin/approve [post]
func (AgentApi) Approve(c *gin.Context) {
	adminID := utils.GetUserID(c)
	var req dxapiReq.AgentApproveReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	agent, err := agentService.AgentServices.Approve(adminID, req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(agent, c)
}

// Reject
// @Tags Agent
// @Summary Admin rejects a pending agent with a reason
// @Produce application/json
// @Param data body dxapiReq.AgentRejectReq true "reject payload"
// @Success 200 {object} response.Response
// @Router /api/v1/agent/admin/reject [post]
func (AgentApi) Reject(c *gin.Context) {
	adminID := utils.GetUserID(c)
	var req dxapiReq.AgentRejectReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	agent, err := agentService.AgentServices.Reject(adminID, req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(agent, c)
}

// Suspend
// @Tags Agent
// @Summary Admin suspends or reactivates an agent
// @Produce application/json
// @Param data body dxapiReq.AgentSuspendReq true "suspend payload"
// @Success 200 {object} response.Response
// @Router /api/v1/agent/admin/suspend [post]
func (AgentApi) Suspend(c *gin.Context) {
	adminID := utils.GetUserID(c)
	var req dxapiReq.AgentSuspendReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	agent, err := agentService.AgentServices.Suspend(adminID, req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(agent, c)
}

// AddNote
// @Tags Agent
// @Summary Admin adds an internal note on an agent
// @Produce application/json
// @Param data body dxapiReq.AgentNoteReq true "note payload"
// @Success 200 {object} response.Response
// @Router /api/v1/agent/admin/note [post]
func (AgentApi) AddNote(c *gin.Context) {
	adminID := utils.GetUserID(c)
	var req dxapiReq.AgentNoteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	note, err := agentService.AgentServices.AddNote(adminID, req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(note, c)
}
