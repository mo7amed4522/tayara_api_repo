package dxapi

// AgentPlan commission plan definitions (Basic/Premium/Custom). Reference table: bare ID.
type AgentPlan struct {
	ID                       uint   `json:"id" gorm:"primarykey;comment:ID"`
	Name                     string `json:"name" form:"name" gorm:"size:64;uniqueIndex;comment:Plan name"`
	DefaultCommissionPercent int64  `json:"defaultCommissionPercent" form:"defaultCommissionPercent" gorm:"default:0;comment:Basis points, 10000 means 100 percent"`
	Description              string `json:"description" form:"description" gorm:"size:500;comment:Description"`
	IsActive                 bool   `json:"isActive" form:"isActive" gorm:"default:true;comment:Is active"`
}

func (AgentPlan) TableName() string { return "dx_agent_plan" }
