package catalog

import (
	"time"

	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/model/common"
)

// Package a trip/package listing owned by an agent (inbound UAE or outbound international).
// Money is int64 minor units (fils); DiscountPercent is int64 basis points (10000 == 100%).
type Package struct {
	ID                   uint           `json:"id" gorm:"primarykey;comment:ID"`
	PackageCode          string         `json:"packageCode" form:"packageCode" gorm:"size:64;uniqueIndex;comment:PK- prefixed code"`
	AgentID              uint           `json:"agentId" form:"agentId" gorm:"index;comment:Owning agent ID"`
	Type                 string         `json:"type" form:"type" gorm:"size:32;index;comment:inbound|outbound"`
	Name                 string         `json:"name" form:"name" gorm:"size:191;index;comment:Package name"`
	CoverImageURL        string         `json:"coverImageURL" form:"coverImageURL" gorm:"size:500;comment:Cover image URL"`
	Description          string         `json:"description" form:"description" gorm:"size:1000;comment:Description"`
	GalleryImages        common.JSONMap `json:"galleryImages" form:"galleryImages" gorm:"comment:Image gallery"`
	DestinationCountryID uint           `json:"destinationCountryId" form:"destinationCountryId" gorm:"index;comment:Destination country ID"`
	Location             string         `json:"location" form:"location" gorm:"size:128;comment:Location (UAE for inbound)"`
	Category             string         `json:"category" form:"category" gorm:"size:64;index;comment:Category code (honeymoon, adventure, ...)"`
	DaysDuration         int            `json:"daysDuration" form:"daysDuration" gorm:"comment:Days duration"`
	AvailabilityType     string         `json:"availabilityType" form:"availabilityType" gorm:"size:32;default:open_calendar;comment:open_calendar|fixed_dates"`
	DisplayPrice         int64          `json:"displayPrice" form:"displayPrice" gorm:"default:0;comment:Display price, minor units"`
	EstimatedPrice       int64          `json:"estimatedPrice" form:"estimatedPrice" gorm:"default:0;comment:Estimated price, minor units"`
	DiscountPercent      int64          `json:"discountPercent" form:"discountPercent" gorm:"default:0;comment:Basis points, min 300 (3%)"`
	Status               string         `json:"status" form:"status" gorm:"size:32;index;default:draft;comment:Package status"`
	FlagIssueType        string         `json:"flagIssueType" form:"flagIssueType" gorm:"size:64;comment:Flag issue type"`
	FlagDetails          string         `json:"flagDetails" form:"flagDetails" gorm:"size:500;comment:Flag details"`
	AdminNote            string         `json:"adminNote" form:"adminNote" gorm:"size:500;comment:Admin note"`
	SubmittedAt          *time.Time     `json:"submittedAt" form:"submittedAt" gorm:"comment:Submitted for approval at"`
	ApprovedBy           *uint          `json:"approvedBy" form:"approvedBy" gorm:"index;comment:Approving admin ID"`
	ApprovedAt           *time.Time     `json:"approvedAt" form:"approvedAt" gorm:"comment:Approved at"`
	global.GVA_MODEL
}

func (Package) TableName() string { return "catalog_package" }
