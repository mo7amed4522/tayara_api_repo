package dxapi

import "github.com/mo7amed4522/tayara_api/global"

// AgentDocument license/verification files an agent uploads at registration.
type AgentDocument struct {
	ID           uint   `json:"id" gorm:"primarykey;comment:ID"`
	AgentID      uint   `json:"agentId" form:"agentId" gorm:"index;comment:Agent ID"`
	DocType      string `json:"docType" form:"docType" gorm:"size:64;index;comment:Document type (license/pdf/image)"`
	FileURL      string `json:"fileUrl" form:"fileUrl" gorm:"size:500;comment:File URL"`
	OriginalName string `json:"originalName" form:"originalName" gorm:"size:191;comment:Original file name"`
	MIME         string `json:"mime" form:"mime" gorm:"size:128;comment:MIME type"`
	Status       string `json:"status" form:"status" gorm:"size:32;index;default:pending;comment:Verification status"`
	global.GVA_MODEL
}

func (AgentDocument) TableName() string { return "dx_agent_document" }
