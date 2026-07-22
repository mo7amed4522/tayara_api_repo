package global

import (
	"time"

	"gorm.io/gorm"
)

type GVA_MODEL struct {
	CreatedAt time.Time      // Created time
	UpdatedAt time.Time      // Updated time
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"` // Deleted time
}
