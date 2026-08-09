package dxapi

import (
	"errors"
	"time"

	"github.com/mo7amed4522/tayara_api/constants"
	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/model/dxapi"
	dxapiReq "github.com/mo7amed4522/tayara_api/model/dxapi/request"
	dxapiResp "github.com/mo7amed4522/tayara_api/model/dxapi/response"
	"github.com/mo7amed4522/tayara_api/utils"
)

type AgentServices struct{}

// agentLoginAdapter lets utils.LoginToken treat an agent as the system.Login interface
// (the agent model exposes GetAgentname/GetAgentId instead of the system method names).
type agentLoginAdapter struct{ agent *dxapi.DxAgentProfile }

func (a agentLoginAdapter) GetUsername() string { return a.agent.GetAgentname() }
func (a agentLoginAdapter) GetNickname() string { return a.agent.CompanyName }
func (a agentLoginAdapter) GetUserId() uint     { return a.agent.ID }
func (a agentLoginAdapter) GetUserInfo() any    { return a.agent }

// Register creates a pending-approval agent account (public).
func (s *AgentServices) Register(req dxapiReq.AgentRegisterReq) (*dxapi.DxAgentProfile, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	var exists dxapi.DxAgentProfile
	if err := global.GVA_DB.Where("business_email = ?", req.BusinessEmail).First(&exists).Error; err == nil {
		return nil, errors.New("business email already registered")
	}
	planID := req.PlanID
	agent := &dxapi.DxAgentProfile{
		CompanyName:      req.CompanyName,
		ResponsibleName:  req.ResponsibleName,
		BusinessEmail:    req.BusinessEmail,
		Password:         utils.BcryptHash(req.Password),
		PhoneCountryCode: req.PhoneCountryCode,
		Phone:            req.Phone,
		Country:          req.Country,
		CompanyAddress:   req.CompanyAddress,
		LicenseNumber:    req.LicenseNumber,
		PlanID:           planID,
		Status:           constants.AgentStatusPending,
	}
	if err := global.GVA_DB.Create(agent).Error; err != nil {
		return nil, err
	}
	return agent, nil
}

// Login authenticates an agent and issues the Manage (A-token) JWT.
func (s *AgentServices) Login(req dxapiReq.AgentLoginReq) (*dxapiResp.AgentLoginResponse, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	var agent dxapi.DxAgentProfile
	if err := global.GVA_DB.Where("business_email = ?", req.BusinessEmail).First(&agent).Error; err != nil {
		return nil, errors.New("invalid email or password")
	}
	if !utils.BcryptCheck(req.Password, agent.Password) {
		return nil, errors.New("invalid email or password")
	}
	if agent.Status != constants.AgentStatusActive {
		return nil, errors.New("agent account is not active (pending approval / suspended / rejected)")
	}
	adapter := agentLoginAdapter{agent: &agent}
	token, claims, err := utils.LoginTokenWithAuthority(adapter, constants.AuthorityAgent)
	if err != nil {
		return nil, err
	}
	// persist the last-login timestamp
	now := time.Now()
	_ = global.GVA_DB.Model(&dxapi.DxAgentProfile{}).Where("id = ?", agent.ID).
		Update("last_login_at", &now).Error
	return &dxapiResp.AgentLoginResponse{
		Agent:     agent,
		Token:     token,
		ExpiresAt: claims.RegisteredClaims.ExpiresAt.Unix() * 1000,
	}, nil
}

// GetProfile returns an agent's own profile by id.
func (s *AgentServices) GetProfile(agentID uint) (*dxapi.DxAgentProfile, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	var agent dxapi.DxAgentProfile
	if err := global.GVA_DB.First(&agent, agentID).Error; err != nil {
		return nil, errors.New("agent not found")
	}
	return &agent, nil
}

// ChangePassword updates the agent's own password after verifying the old one.
func (s *AgentServices) ChangePassword(agentID uint, req dxapiReq.ChangeAgentPasswordReq) error {
	if global.GVA_DB == nil {
		return errors.New("db not init")
	}
	var agent dxapi.DxAgentProfile
	if err := global.GVA_DB.First(&agent, agentID).Error; err != nil {
		return errors.New("agent not found")
	}
	if !utils.BcryptCheck(req.OldPassword, agent.Password) {
		return errors.New("old password is incorrect")
	}
	return global.GVA_DB.Model(&dxapi.DxAgentProfile{}).Where("id = ?", agentID).
		Update("password", utils.BcryptHash(req.NewPassword)).Error
}

// --- admin actions ---

// Approve approves a pending agent and sets commission % + SLA hours at approval.
func (s *AgentServices) Approve(adminID uint, req dxapiReq.AgentApproveReq) (*dxapi.DxAgentProfile, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	var agent dxapi.DxAgentProfile
	if err := global.GVA_DB.First(&agent, req.AgentID).Error; err != nil {
		return nil, errors.New("agent not found")
	}
	if agent.Status != constants.AgentStatusPending {
		return nil, errors.New("only pending agents can be approved")
	}
	now := time.Now()
	commission := req.CommissionPercent
	if commission == 0 {
		// default from the agent's plan
		if agent.PlanID != nil {
			var plan dxapi.AgentPlan
			if err := global.GVA_DB.First(&plan, *agent.PlanID).Error; err == nil {
				commission = plan.DefaultCommissionPercent
			}
		}
	}
	updates := map[string]interface{}{
		"status": constants.AgentStatusActive, "approved_by": adminID, "approved_at": &now,
	}
	if commission > 0 {
		updates["commission_percent"] = commission
	}
	if req.SLAHours > 0 {
		updates["sla_hours"] = req.SLAHours
	}
	if err := global.GVA_DB.Model(&dxapi.DxAgentProfile{}).Where("id = ?", agent.ID).Updates(updates).Error; err != nil {
		return nil, err
	}
	agent.Status = constants.AgentStatusActive
	return &agent, nil
}

// Reject marks a pending agent rejected with a reason.
func (s *AgentServices) Reject(adminID uint, req dxapiReq.AgentRejectReq) (*dxapi.DxAgentProfile, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	var agent dxapi.DxAgentProfile
	if err := global.GVA_DB.First(&agent, req.AgentID).Error; err != nil {
		return nil, errors.New("agent not found")
	}
	now := time.Now()
	if err := global.GVA_DB.Model(&dxapi.DxAgentProfile{}).Where("id = ?", agent.ID).Updates(map[string]interface{}{
		"status": constants.AgentStatusRejected, "rejected_reason": req.Reason,
		"approved_by": adminID, "approved_at": &now,
	}).Error; err != nil {
		return nil, err
	}
	agent.Status = constants.AgentStatusRejected
	return &agent, nil
}

// Suspend suspends (or reactivates) an active/suspended agent.
func (s *AgentServices) Suspend(adminID uint, req dxapiReq.AgentSuspendReq) (*dxapi.DxAgentProfile, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	var agent dxapi.DxAgentProfile
	if err := global.GVA_DB.First(&agent, req.AgentID).Error; err != nil {
		return nil, errors.New("agent not found")
	}
	target := constants.AgentStatusSuspended
	if agent.Status == constants.AgentStatusSuspended {
		target = constants.AgentStatusActive
	}
	if err := global.GVA_DB.Model(&dxapi.DxAgentProfile{}).Where("id = ?", agent.ID).Updates(map[string]interface{}{
		"status": target, "approved_by": adminID,
	}).Error; err != nil {
		return nil, err
	}
	agent.Status = target
	return &agent, nil
}

// AddNote appends an internal admin note to an agent.
func (s *AgentServices) AddNote(adminID uint, req dxapiReq.AgentNoteReq) (*dxapi.AgentNote, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	note := &dxapi.AgentNote{AgentID: req.AgentID, AdminID: adminID, NoteText: req.NoteText}
	if err := global.GVA_DB.Create(note).Error; err != nil {
		return nil, err
	}
	return note, nil
}

// List returns agents (admin: all; scoped by status filter).
func (s *AgentServices) List(req dxapiReq.AgentPageReq) ([]dxapi.DxAgentProfile, int64, int64, error) {
	if global.GVA_DB == nil {
		return nil, 0, 0, errors.New("db not init")
	}
	db := global.GVA_DB.Model(&dxapi.DxAgentProfile{})
	if req.Status != "" {
		db = db.Where("status = ?", req.Status)
	}
	var list []dxapi.DxAgentProfile
	pr, err := utils.PaginateWithCustomQuery(db, db.Order("created_at desc"), req.Page, req.PageSize, &list)
	if err != nil {
		return nil, 0, 0, err
	}
	return list, pr.Total, pr.TotalPage, nil
}

// GetAdminDetail returns an agent with its documents + notes (admin view).
func (s *AgentServices) GetAdminDetail(agentID uint) (*dxapiResp.AgentDetail, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	var agent dxapi.DxAgentProfile
	if err := global.GVA_DB.First(&agent, agentID).Error; err != nil {
		return nil, errors.New("agent not found")
	}
	var documents []dxapi.AgentDocument
	var notes []dxapi.AgentNote
	_ = global.GVA_DB.Where("agent_id = ?", agentID).Order("created_at desc").Find(&documents).Error
	_ = global.GVA_DB.Where("agent_id = ?", agentID).Order("created_at desc").Find(&notes).Error
	return &dxapiResp.AgentDetail{DxAgentProfile: agent, Documents: documents, Notes: notes}, nil
}
