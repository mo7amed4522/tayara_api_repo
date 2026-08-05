package engage

import "github.com/mo7amed4522/tayara_api/global"

// Review a post-completion review & rating of a booking (optionally linked to an agent/package).
type Review struct {
	ID           uint   `json:"id" gorm:"primarykey;comment:ID"`
	BookingID    uint   `json:"bookingId" form:"bookingId" gorm:"index;comment:Booking ID"`
	UserID       uint   `json:"userId" form:"userId" gorm:"index;comment:Authoring user ID"`
	AgentID      *uint  `json:"agentId" form:"agentId" gorm:"index;comment:Reviewed agent ID"`
	PackageID    *uint  `json:"packageId" form:"packageId" gorm:"index;comment:Reviewed package ID"`
	Rating       int    `json:"rating" form:"rating" gorm:"comment:Rating 1-5"`
	ReviewText   string `json:"reviewText" form:"reviewText" gorm:"size:2000;comment:Review text"`
	ReviewStatus string `json:"reviewStatus" form:"reviewStatus" gorm:"size:16;index;default:active;comment:active|disabled"`
	global.GVA_MODEL
}

func (Review) TableName() string { return "engage_review" }
