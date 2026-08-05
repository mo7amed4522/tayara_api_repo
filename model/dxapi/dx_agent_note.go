package dxapi

import "github.com/mo7amed4522/tayara_api/global"

// AgentNote internal admin notes on an agent (Add Note action during approval/monitoring).
type AgentNote struct {
	ID       uint   `json:"id" gorm:"primarykey;comment:ID"`
	AgentID  uint   `json:"agentId" form:"agentId" gorm:"index;comment:Agent ID"`
	AdminID  uint   `json:"adminId" form:"adminId" gorm:"index;comment:Authoring admin ID"`
	NoteText string `json:"noteText" form:"noteText" gorm:"size:1000;comment:Note text"`
	global.GVA_MODEL
}

func (AgentNote) TableName() string { return "dx_agent_note" }
