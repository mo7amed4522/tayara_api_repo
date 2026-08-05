package dxapi

import (
	"time"

	"github.com/mo7amed4522/tayara_api/global"
)

type Login interface {
	GetAgentname() string
	GetResponsibleName() string
	GetAgentId() uint
	GetUserInfo() any
}

var _ Login = new(DxAgentProfile)

type DxAgentProfile struct {
	ID                uint       `json:"id" gorm:"primarykey;comment:id of admin"` // UUID
	CompanyName       string     `json:"companyName" form:"companyName" gorm:"size:191;index;comment:Company name"`
	Password          string     `json:"-"  gorm:"comment:Password"` // Password
	ResponsibleName   string     `json:"responsibleName" form:"responsibleName" gorm:"size:191;comment:Agent responsible person"`
	HeaderImg         string     `json:"headerImg" gorm:"default:https://raw.githubusercontent.com/mo7amed4522/new_app/refs/heads/master/logo.png"` // HeaderImg
	PhoneCountryCode  string     `json:"phoneCountryCode" gorm:"comment:User phone number country code"`                                            // User phone number country code
	Phone             string     `json:"phone"  gorm:"comment:Phone"`                                                                               // Phone
	BusinessEmail     string     `json:"businessEmail" form:"businessEmail" gorm:"size:191;uniqueIndex;comment:Business email"`
	Country           string     `json:"country" form:"country" gorm:"size:128;index;comment:Country"`
	CompanyAddress    string     `json:"companyAddress" form:"companyAddress" gorm:"size:500;comment:Company address"`
	LicenseNumber     string     `json:"licenseNumber" form:"licenseNumber" gorm:"size:128;index;comment:License number"`
	Status            string     `json:"status" form:"status" gorm:"size:32;index;default:pending;comment:Agent status"`
	PlanID            *uint      `json:"planId" form:"planId" gorm:"index;comment:Commission plan ID"`
	ShortDescription  *string    `json:"shortDescription" form:"shortDescription" gorm:"size:200;comment:Company address"`
	LongDescription   *string    `json:"longDescription" form:"longDescription" gorm:"size:500;comment:Company address"`
	CommissionPercent int64      `json:"commissionPercent" form:"commissionPercent" gorm:"default:0;comment:Basis points, 10000 means 100 percent"`
	SLAHours          int        `json:"slaHours" form:"slaHours" gorm:"default:2;comment:Agent response SLA hours"`
	ApprovedBy        *uint      `json:"approvedBy" form:"approvedBy" gorm:"index;comment:Approving admin user ID"`
	RejectedReason    string     `json:"rejectedReason" form:"rejectedReason" gorm:"size:500;comment:Rejection reason"`
	ApprovedAt        *time.Time `json:"approvedAt" form:"approvedAt" gorm:"comment:Approved at"`
	LastLoginAt       *time.Time `json:"lastLoginAt" form:"lastLoginAt" gorm:"comment:Last login at"`
	global.GVA_MODEL
}

func (DxAgentProfile) TableName() string {
	return "dx_agent_profile"
}

func (s *DxAgentProfile) GetAgentname() string {
	return s.CompanyName
}

func (s *DxAgentProfile) GetResponsibleName() string {
	return s.ResponsibleName
}

func (s *DxAgentProfile) GetAgentId() uint {
	return s.ID
}

func (s *DxAgentProfile) GetUserInfo() any {
	return *s
}
