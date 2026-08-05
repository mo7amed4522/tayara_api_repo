package base

// Nationality passport nationalities (drives outbound visa-waiver logic).
type Nationality struct {
	ID       uint   `json:"id" gorm:"primarykey;comment:ID"`
	Code     string `json:"code" form:"code" gorm:"size:32;index;comment:ISO nationality code"`
	Name     string `json:"name" form:"name" gorm:"size:128;index;comment:Nationality name"`
	IsActive bool   `json:"isActive" form:"isActive" gorm:"default:true;comment:Is active"`
}

func (Nationality) TableName() string { return "base_nationality" }
