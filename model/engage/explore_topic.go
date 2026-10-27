package engage

import (
	"time"

	"github.com/mo7amed4522/tayara_api/global"
	"gorm.io/gorm"
)

// ExploreTopic represents a curated topic/collection for the Explore page
// e.g., "Adventure Trips", "Luxury Escapes", "Cultural Journeys", etc.
type ExploreTopic struct {
	ID          uint       `json:"id" gorm:"primarykey;comment:ID"`
	Code        string     `json:"code" form:"code" gorm:"size:64;uniqueIndex;comment:Unique code (adventure, luxury, cultural, honeymoon, family, friends)"`
	Name        string     `json:"name" form:"name" gorm:"size:128;index;comment:Display name (English)"`
	NameAr      string     `json:"nameAr" form:"nameAr" gorm:"size:128;comment:Display name (Arabic)"`
	Description string     `json:"description" form:"description" gorm:"size:500;comment:Short description"`
	ImageURL    string     `json:"imageUrl" form:"imageUrl" gorm:"size:500;comment:Topic banner/hero image URL"`
	IconURL     string     `json:"iconUrl" form:"iconUrl" gorm:"size:500;comment:Topic icon URL"`
	SortOrder   int        `json:"sortOrder" form:"sortOrder" gorm:"default:0;comment:Display order"`
	IsActive    bool       `json:"isActive" form:"isActive" gorm:"default:true;comment:Visible to users"`
	CreatedAt   time.Time  `json:"createdAt" gorm:"comment:Creation time"`
	UpdatedAt   time.Time  `json:"updatedAt" gorm:"comment:Last update time"`
	DeletedAt   *time.Time `json:"-" gorm:"index;comment:Soft delete timestamp"`

	// Computed fields (not stored in DB)
	PackageCount int `json:"packageCount" gorm:"-;comment:Number of packages in this topic"`
}

func (ExploreTopic) TableName() string { return "engage_explore_topic" }

// PackageTopic represents the many-to-many relationship between packages and explore topics
type PackageTopic struct {
	PackageID uint      `json:"packageId" gorm:"primaryKey;index;comment:Package ID"`
	TopicID   uint      `json:"topicId" gorm:"primaryKey;index;comment:Topic ID"`
	AddedBy   uint      `json:"addedBy" gorm:"comment:Admin/Agent ID who added this association"`
	AddedAt   time.Time `json:"addedAt" gorm:"autoCreateTime;comment:Association creation time"`

	// Relations
	Package *ExploreTopic `json:"-" gorm:"foreignKey:PackageID;references:ID"`
	Topic   *ExploreTopic `json:"-" gorm:"foreignKey:TopicID;references:ID"`
}

func (PackageTopic) TableName() string { return "engage_package_topic" }

// ExploreTopicWithPackages is used for queries that join topics with packages
type ExploreTopicWithPackages struct {
	ExploreTopic
	Packages []PackageTopic `json:"packages" gorm:"foreignKey:TopicID;references:ID"`
}

// BeforeCreate hook to set defaults
func (t *ExploreTopic) BeforeCreate(tx *gorm.DB) error {
	if t.SortOrder == 0 {
		// Auto-assign sort order at the end
		var maxOrder int
		tx.Model(&ExploreTopic{}).Select("COALESCE(MAX(sort_order), 0)").Scan(&maxOrder)
		t.SortOrder = maxOrder + 1
	}
	return nil
}

// GetActiveTopics returns all active topics ordered by sort_order
func GetActiveTopics() ([]ExploreTopic, error) {
	var topics []ExploreTopic
	err := global.GVA_DB.Where("is_active = ? AND deleted_at IS NULL", true).
		Order("sort_order asc, id asc").
		Find(&topics).Error
	return topics, err
}

// GetTopicByCode returns a topic by its code
func GetTopicByCode(code string) (*ExploreTopic, error) {
	var topic ExploreTopic
	err := global.GVA_DB.Where("code = ? AND deleted_at IS NULL", code).First(&topic).Error
	if err != nil {
		return nil, err
	}
	return &topic, nil
}

// GetTopicByID returns a topic by its ID
func GetTopicByID(id uint) (*ExploreTopic, error) {
	var topic ExploreTopic
	err := global.GVA_DB.Where("id = ? AND deleted_at IS NULL", id).First(&topic).Error
	if err != nil {
		return nil, err
	}
	return &topic, nil
}
