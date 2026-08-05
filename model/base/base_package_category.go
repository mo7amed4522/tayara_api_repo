package base

// PackageCategory outbound trip classification lookup (honeymoon, adventure, ...).
type PackageCategory struct {
	ID          uint   `json:"id" gorm:"primarykey;comment:ID"`
	Code        string `json:"code" form:"code" gorm:"size:64;uniqueIndex;comment:Category code"`
	Name        string `json:"name" form:"name" gorm:"size:128;index;comment:Display name"`
	Description string `json:"description" form:"description" gorm:"size:500;comment:Description"`
	SortOrder   int    `json:"sortOrder" form:"sortOrder" gorm:"default:0;comment:Sort order"`
	IsActive    bool   `json:"isActive" form:"isActive" gorm:"default:true;comment:Is active"`
}

func (PackageCategory) TableName() string { return "base_package_category" }
