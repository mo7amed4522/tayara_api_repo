package base

// DocumentType platform document library that agents/required-profiles reference
// (e.g. passport_copy, emirates_id, personal_photo).
type DocumentType struct {
	ID                 uint   `json:"id" gorm:"primarykey;comment:ID"`
	Code               string `json:"code" form:"code" gorm:"size:64;uniqueIndex;comment:Machine code"`
	Name               string `json:"name" form:"name" gorm:"size:128;index;comment:Display name"`
	IsMandatoryDefault bool   `json:"isMandatoryDefault" form:"isMandatoryDefault" gorm:"default:false;comment:Required by default"`
	Description        string `json:"description" form:"description" gorm:"size:500;comment:Description"`
	IsActive           bool   `json:"isActive" form:"isActive" gorm:"default:true;comment:Is active"`
}

func (DocumentType) TableName() string { return "base_document_type" }
