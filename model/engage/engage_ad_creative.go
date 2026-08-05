package engage

import (
	"time"

	"github.com/mo7amed4522/tayara_api/global"
)

// AdCreative a promotional banner (Home Top Banner / Explore Top Banner). The scheduler flips
// scheduled -> active at start_datetime and active -> expired at end_datetime.
type AdCreative struct {
	ID              uint       `json:"id" gorm:"primarykey;comment:ID"`
	AdCode          string     `json:"adCode" form:"adCode" gorm:"size:64;uniqueIndex;comment:AD- prefixed code"`
	Title           string     `json:"title" form:"title" gorm:"size:191;comment:Title"`
	Placement       string     `json:"placement" form:"placement" gorm:"size:32;index;comment:home_top_banner|explore_top_banner"`
	Status          string     `json:"status" form:"status" gorm:"size:32;index;default:draft;comment:draft|scheduled|active|disabled|expired"`
	DisplayOrder    int        `json:"displayOrder" form:"displayOrder" gorm:"default:0;comment:Order within placement"`
	StartDatetime   *time.Time `json:"startDatetime" form:"startDatetime" gorm:"comment:Schedule start"`
	EndDatetime     *time.Time `json:"endDatetime" form:"endDatetime" gorm:"comment:Schedule end"`
	BannerImageURL  string     `json:"bannerImageURL" form:"bannerImageURL" gorm:"size:500;comment:Banner image URL"`
	OverlayTitle    string     `json:"overlayTitle" form:"overlayTitle" gorm:"size:191;comment:Overlay title"`
	OverlaySubtitle string     `json:"overlaySubtitle" form:"overlaySubtitle" gorm:"size:191;comment:Overlay subtitle"`
	CtaLabel        string     `json:"ctaLabel" form:"ctaLabel" gorm:"size:64;comment:CTA label"`
	RedirectType    string     `json:"redirectType" form:"redirectType" gorm:"size:32;index;comment:packages_by_destination|agents_by_destination|single_agent_packages"`
	TargetCountryID *uint      `json:"targetCountryId" form:"targetCountryId" gorm:"index;comment:Target country ID"`
	TargetAgentID   *uint      `json:"targetAgentId" form:"targetAgentId" gorm:"index;comment:Target agent ID"`
	TargetPackageID *uint      `json:"targetPackageId" form:"targetPackageId" gorm:"index;comment:Target package ID"`
	Impressions     int64      `json:"impressions" form:"impressions" gorm:"default:0;comment:Impression count"`
	Clicks          int64      `json:"clicks" form:"clicks" gorm:"default:0;comment:Click count"`
	CreatedBy       uint       `json:"createdBy" form:"createdBy" gorm:"index;comment:Creating admin ID"`
	global.GVA_MODEL
}

func (AdCreative) TableName() string { return "engage_ad_creative" }
