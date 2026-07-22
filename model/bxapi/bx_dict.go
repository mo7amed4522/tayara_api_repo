package bxapi

import (
	"gorm.io/gorm"
)

// BxDict Dictionary table model
type BxDict struct {
	ID          uint   `json:"id" gorm:"primaryKey;autoIncrement;comment:ID"`                                // ID
	Value       string `json:"value" gorm:"type:varchar(100);not null;default:'';comment:Value"`             // Value
	Label       string `json:"label" gorm:"type:varchar(100);not null;default:'';comment:Label"`             // Label
	Type        string `json:"type" gorm:"type:varchar(100);not null;default:'';comment:Type"`               // Type
	Description string `json:"description" gorm:"type:varchar(100);not null;default:'';comment:Description"` // Description
	Sort        uint   `json:"sort" gorm:"not null;default:0;comment:Sort"`                                  // Sort
	ParentId    uint   `json:"parentId" gorm:"not null;default:0;comment:ParentId"`                          // ParentId
	Remarks     string `json:"remarks" gorm:"type:varchar(255);not null;default:'';comment:Remarks"`         // Remarks
	IsDeleted   bool   `json:"isDeleted" gorm:"not null;default:0;comment:IsDeleted"`                        // IsDeleted
}

func (BxDict) TableName() string {
	return "bx_dict"
}

// BeforeCreate BeforeCreate
func (d *BxDict) BeforeCreate(tx *gorm.DB) error {
	return nil
}

// BeforeUpdate BeforeUpdate
func (d *BxDict) BeforeUpdate(tx *gorm.DB) error {
	return nil
}
