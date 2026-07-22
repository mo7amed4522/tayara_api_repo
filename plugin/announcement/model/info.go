package model

import (
	"github.com/mo7amed4522/tayara_api/global"
	"gorm.io/datatypes"
)

// Info announcement structure
type Info struct {
	global.GVA_MODEL
	Title       string         `json:"title" form:"title" gorm:"column:title;comment:Announcement title	;"`                                      // Title
	Content     string         `json:"content" form:"content" gorm:"column:content;comment:Announcement content;type:text;"`                     // Content
	UserID      *int           `json:"userID" form:"userID" gorm:"column:user_id;comment:Author;"`                                               // Author
	Attachments datatypes.JSON `json:"attachments" form:"attachments" gorm:"column:attachments;comment:Attachments;" swaggertype:"array,object"` // Attachments
}

// TableName announcement Info custom table name gva_announcements_info
func (Info) TableName() string {
	return "gva_announcements_info"
}
