package catalog

import "github.com/mo7amed4522/tayara_api/global"

// ItineraryDay a day in the package daily itinerary.
type ItineraryDay struct {
	ID          uint               `json:"id" gorm:"primarykey;comment:ID"`
	PackageID   uint               `json:"packageId" form:"packageId" gorm:"index;comment:Package ID"`
	DayNumber   int                `json:"dayNumber" form:"dayNumber" gorm:"comment:Day number (1-based)"`
	Title       string             `json:"title" form:"title" gorm:"size:191;comment:Day title, e.g. Day1: Luxor Embarkation"`
	Description string             `json:"description" form:"description" gorm:"size:2000;comment:Day-level summary"`
	SortOrder   int                `json:"sortOrder" form:"sortOrder" gorm:"default:0;comment:Sort order"`
	Items       []ItineraryItem    `json:"items" gorm:"-"`
	global.GVA_MODEL
}

func (ItineraryDay) TableName() string { return "catalog_itinerary_day" }

// ItineraryItem a time-slotted entry inside an ItineraryDay (e.g. "12:00: Check-in").
type ItineraryItem struct {
	ID          uint   `json:"id" gorm:"primarykey;comment:ID"`
	ItineraryDayID uint `json:"itineraryDayId" form:"itineraryDayId" gorm:"index;comment:Parent day ID"`
	Time        string `json:"time" form:"time" gorm:"size:32;comment:Time label, e.g. 12:00"`
	Title       string `json:"title" form:"title" gorm:"size:191;comment:Entry title"`
	Description string `json:"description" form:"description" gorm:"size:1000;comment:Entry description"`
	SortOrder   int    `json:"sortOrder" form:"sortOrder" gorm:"default:0;comment:Sort order within the day"`
	global.GVA_MODEL
}

func (ItineraryItem) TableName() string { return "catalog_itinerary_item" }
