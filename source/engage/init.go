// Package engage seeds notifications + templates, reviews, ad creatives/clicks, wishlist,
// shares, coupons and device tokens so the content/engagement surfaces have realistic data.
package engage

import (
	"context"
	"time"

	engmodel "github.com/mo7amed4522/tayara_api/model/engage"
	"github.com/mo7amed4522/tayara_api/service/system"
	"gorm.io/gorm"
)

const initOrderEngage = system.InitOrderExternal + 8

type engageInit struct{}

func init() {
	system.RegisterInit(initOrderEngage, &engageInit{})
}

func (e *engageInit) InitializerName() string { return "demo_engage" }

// MigrateTable/TableCreated are no-ops: tables migrate via gorm_biz AutoMigrate.
func (e *engageInit) MigrateTable(ctx context.Context) (context.Context, error) { return ctx, nil }
func (e *engageInit) TableCreated(ctx context.Context) bool                     { return true }

func (e *engageInit) DataInserted(ctx context.Context) bool {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok || db == nil {
		return false
	}
	var n int64
	return db.Model(&engmodel.AdCreative{}).Count(&n).Error == nil && n > 0 &&
		db.Model(&engmodel.ExploreTopic{}).Count(&n).Error == nil && n > 0
}

func (e *engageInit) InitializeData(ctx context.Context) (context.Context, error) {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok || db == nil {
		return ctx, system.ErrMissingDBContext
	}
	now := time.Now()
	uAhmed := userID(db, "ahmed.ali")
	uOmar := userID(db, "omar.h")
	pIstanbul := packageID(db, "PK-1001")
	pDubai := packageID(db, "PK-1002")
	pBali := packageID(db, "PK-1003")
	pMaldives := packageID(db, "PK-1004")
	bk1 := bookingID(db, "BK-2001")

	if err := db.Create(seedNotificationTemplates()).Error; err != nil {
		return ctx, err
	}
	if err := db.Create(seedNotifications(now, uAhmed, uOmar)).Error; err != nil {
		return ctx, err
	}
	if err := db.Create(seedReviews(bk1, uAhmed, pIstanbul)).Error; err != nil {
		return ctx, err
	}
	if err := db.Create(seedAds(db, now)).Error; err != nil {
		return ctx, err
	}
	if err := db.Create(seedExploreTopics(db, now)).Error; err != nil {
		return ctx, err
	}
	adID := adByCode(db, "AD-5001")
	if err := db.Create(&[]engmodel.AdClick{
		{AdID: adID, UserID: uintPtr(uAhmed), SourcePage: "home"},
		{AdID: adID, UserID: uintPtr(uOmar), SourcePage: "home"},
	}).Error; err != nil {
		return ctx, err
	}
	if err := db.Create(&[]engmodel.Wishlist{
		{UserID: uAhmed, PackageID: pIstanbul},
		{UserID: uAhmed, PackageID: pDubai},
		{UserID: uOmar, PackageID: pIstanbul},
	}).Error; err != nil {
		return ctx, err
	}
	if err := db.Create(&[]engmodel.Share{
		{UserID: uAhmed, PackageID: pIstanbul, Channel: "whatsapp", SharedAt: tAdd(now, -3*24)},
		{UserID: uOmar, PackageID: pDubai, Channel: "instagram", SharedAt: tAdd(now, -1*24)},
	}).Error; err != nil {
		return ctx, err
	}
	if err := db.Create(seedCoupons(now, uAhmed)).Error; err != nil {
		return ctx, err
	}
	if err := db.Create(&[]engmodel.DeviceToken{
		{UserID: uAhmed, Token: "fcm-token-ahmed-0001", Platform: "ios", IsActive: true},
		{UserID: uAhmed, Token: "fcm-token-ahmed-0002", Platform: "android", IsActive: true},
		{UserID: uOmar, Token: "fcm-token-omar-0001", Platform: "android", IsActive: true},
	}).Error; err != nil {
		return ctx, err
	}
	// Seed package-topic associations
	if err := db.Create(seedPackageTopics(pIstanbul, pDubai, pBali, pMaldives)).Error; err != nil {
		return ctx, err
	}
	return ctx, nil
}

func seedNotificationTemplates() []engmodel.NotificationTemplate {
	return []engmodel.NotificationTemplate{
		{EventType: "booking_request", Channel: "push", Subject: "New booking request", Body: "A customer requested a new trip.", IsActive: true},
		{EventType: "booking_request", Channel: "email", Subject: "New booking request", Body: "Please review the new booking request.", IsActive: true},
		{EventType: "price_confirmed", Channel: "push", Subject: "Price confirmed", Body: "Your trip price has been confirmed.", IsActive: true},
		{EventType: "payment_success", Channel: "push", Subject: "Payment successful", Body: "Your payment was captured.", IsActive: true},
		{EventType: "offer_expiring", Channel: "email", Subject: "Offer expiring soon", Body: "Your price offer expires soon.", IsActive: true},
		{EventType: "payout", Channel: "push", Subject: "Payout released", Body: "Your payout has been processed.", IsActive: true},
		{EventType: "refund", Channel: "email", Subject: "Refund processed", Body: "Your refund has been issued.", IsActive: true},
	}
}

func seedNotifications(_ time.Time, uAhmed, uOmar uint) []engmodel.Notification {
	return []engmodel.Notification{
		{NotificationCode: "NT-7001", EventType: "booking_request", RecipientType: "agent", RecipientID: 1, Channel: "push", Status: "sent", Title: "New booking request", Body: "BK-2001 requested"},
		{NotificationCode: "NT-7002", EventType: "price_confirmed", RecipientType: "user", RecipientID: uAhmed, Channel: "push", Status: "sent", Title: "Price confirmed", Body: "BK-2001 price confirmed"},
		{NotificationCode: "NT-7003", EventType: "payment_success", RecipientType: "user", RecipientID: uAhmed, Channel: "push", Status: "sent", Title: "Payment successful", Body: "BK-2001 payment captured"},
		{NotificationCode: "NT-7004", EventType: "offer_expiring", RecipientType: "user", RecipientID: uOmar, Channel: "email", Status: "pending", Title: "Offer expiring soon", Body: "BK-2003 offer expires soon"},
		{NotificationCode: "NT-7005", EventType: "payout", RecipientType: "agent", RecipientID: 1, Channel: "push", Status: "sent", Title: "Payout released", Body: "Payout for BK-2001 released"},
		{NotificationCode: "NT-7006", EventType: "refund", RecipientType: "user", RecipientID: uAhmed, Channel: "email", Status: "sent", Title: "Refund processed", Body: "Refund for BK-2006 issued"},
	}
}

func seedReviews(bk1, uAhmed uint, pIstanbul uint) []engmodel.Review {
	agent := uintPtr(1)
	return []engmodel.Review{
		{BookingID: bk1, UserID: uAhmed, AgentID: agent, PackageID: uintPtr(pIstanbul), Rating: 5, ReviewText: "Amazing trip, highly recommend!", ReviewStatus: "active"},
		{BookingID: bk1, UserID: uAhmed, AgentID: agent, PackageID: uintPtr(pIstanbul), Rating: 4, ReviewText: "Great experience overall.", ReviewStatus: "active"},
	}
}

func seedAds(db *gorm.DB, now time.Time) []engmodel.AdCreative {
	activeStart := tAdd(now, -10*24)
	activeEnd := tAdd(now, 20*24)
	schedStart := tAdd(now, 3*24)
	schedEnd := tAdd(now, 30*24)
	expiredStart := tAdd(now, -60*24)
	expiredEnd := tAdd(now, -20*24)
	return []engmodel.AdCreative{
		{
			AdCode: "AD-5001", Title: "Summer in Istanbul", Placement: "home_top_banner", Status: "active",
			DisplayOrder: 1, StartDatetime: activeStart, EndDatetime: activeEnd,
			BannerImageURL: "https://img.magnific.com/free-psd/travel-tourism-social-media-post-template_47987-29099.jpg?semt=ais_hybrid&w=740&q=80",
			OverlayTitle:   "Istanbul", OverlaySubtitle: "From AED 3,200", CtaLabel: "Explore",
			RedirectType:    "packages_by_destination",
			TargetCountryID: uintPtr(countryIDByCode(db, "TR")), Impressions: 1284, Clicks: 96, CreatedBy: 1,
		},
		{
			AdCode: "AD-5002", Title: "Dubai City Break", Placement: "home_top_banner", Status: "active",
			DisplayOrder: 2, StartDatetime: activeStart, EndDatetime: activeEnd,
			BannerImageURL: "https://d1csarkz8obe9u.cloudfront.net/posterpreviews/travel-agency-ad-design-template-b7db8661374e0eb66fad2ff933f59835_screen.jpg?ts=1653886566",
			OverlayTitle:   "Dubai", OverlaySubtitle: "From AED 1,950", CtaLabel: "View deals",
			RedirectType:    "packages_by_destination",
			TargetCountryID: uintPtr(countryIDByCode(db, "AE")), Impressions: 872, Clicks: 41, CreatedBy: 1,
		},
		{
			AdCode: "AD-5003", Title: "Maldives Honeymoon", Placement: "home_top_banner", Status: "active",
			DisplayOrder: 3, StartDatetime: activeStart, EndDatetime: activeEnd,
			BannerImageURL: "https://img.magnific.com/free-psd/tour-travel-agency-instagram-post-social-media-post-web-banner-template-design_84443-25291.jpg?semt=ais_hybrid&w=740&q=80",
			OverlayTitle:   "Maldives", OverlaySubtitle: "Honeymoon specials", CtaLabel: "Book now",
			RedirectType:    "packages_by_destination",
			TargetCountryID: uintPtr(countryIDByCode(db, "MV")), Impressions: 645, Clicks: 58, CreatedBy: 1,
		},
		{
			AdCode: "AD-5004", Title: "Escape to Bali", Placement: "explore_top_banner", Status: "scheduled",
			DisplayOrder: 1, StartDatetime: schedStart, EndDatetime: schedEnd,
			BannerImageURL: "https://img.magnific.com/free-psd/travel-tourism-instagram-post-social-media-post-template_106176-2392.jpg",
			OverlayTitle:   "Bali", OverlaySubtitle: "Adventure awaits", CtaLabel: "Book now",
			RedirectType:    "single_agent_packages",
			TargetPackageID: uintPtr(packageIDByCode(db, "PK-1003")), CreatedBy: 1,
		},
		{
			AdCode: "AD-5005", Title: "Paris Spring Getaway", Placement: "explore_top_banner", Status: "disabled",
			DisplayOrder: 2, StartDatetime: activeStart, EndDatetime: activeEnd,
			BannerImageURL: "https://www.shutterstock.com/image-vector/travel-promo-vector-design-tour-260nw-2172252889.jpg",
			OverlayTitle:   "Paris", OverlaySubtitle: "Cherry blossom season", CtaLabel: "Explore",
			RedirectType:    "packages_by_destination",
			TargetCountryID: uintPtr(countryIDByCode(db, "FR")), Impressions: 120, Clicks: 9, CreatedBy: 2,
		},
		{
			AdCode: "AD-5006", Title: "Bangkok Street Food Fest", Placement: "home_top_banner", Status: "expired",
			DisplayOrder: 4, StartDatetime: expiredStart, EndDatetime: expiredEnd,
			BannerImageURL: "https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcTXRrGOt_cVkV4rlhQMvNUSNXTXDUZq0m1PWPuaeOa3rIyR-ub7nHtZZ4E&s=10",
			OverlayTitle:   "Bangkok", OverlaySubtitle: "Street food festival", CtaLabel: "Explore",
			RedirectType:    "packages_by_destination",
			TargetCountryID: uintPtr(countryIDByCode(db, "TH")), Impressions: 3120, Clicks: 205, CreatedBy: 1,
		},
	}
}

func seedCoupons(now time.Time, uAhmed uint) []engmodel.Coupon {
	return []engmodel.Coupon{
		{Code: "WELCOME100", UserID: uintPtr(uAhmed), Value: 10000, Status: "active", ValidFrom: tAdd(now, -30*24), ValidTo: tAdd(now, 60*24)},
		{Code: "SUMMER50", UserID: nil, Value: 5000, Status: "active", ValidFrom: tAdd(now, -5*24), ValidTo: tAdd(now, 25*24)},
		{Code: "USEDGIFT", UserID: uintPtr(uAhmed), Value: 20000, Status: "used", ValidFrom: tAdd(now, -90*24), ValidTo: tAdd(now, -30*24)},
	}
}

// ---- id lookups ----

func userID(db *gorm.DB, username string) uint {
	var r struct{ ID uint }
	_ = db.Table("bx_user_profile").Select("id").Where("username = ?", username).Scan(&r).Error
	return r.ID
}
func packageID(db *gorm.DB, code string) uint {
	var r struct{ ID uint }
	_ = db.Table("catalog_package").Select("id").Where("package_code = ?", code).Scan(&r).Error
	return r.ID
}
func bookingID(db *gorm.DB, code string) uint {
	var r struct{ ID uint }
	_ = db.Table("bookings").Select("id").Where("booking_code = ?", code).Scan(&r).Error
	return r.ID
}
func adByCode(db *gorm.DB, code string) uint {
	var r struct{ ID uint }
	_ = db.Table("engage_ad_creative").Select("id").Where("ad_code = ?", code).Scan(&r).Error
	return r.ID
}

// countryIDByCode is a shim that resolves a country ID by its ISO code via a helper table call.
func countryIDByCode(db *gorm.DB, code string) uint {
	var r struct{ ID uint }
	_ = db.Table("base_country").Select("id").Where("code = ?", code).Scan(&r).Error
	return r.ID
}
func packageIDByCode(db *gorm.DB, code string) uint {
	var r struct{ ID uint }
	_ = db.Table("catalog_package").Select("id").Where("package_code = ?", code).Scan(&r).Error
	return r.ID
}

// small helpers (package-scoped)
func uintPtr(v uint) *uint { return &v }
func tAdd(t time.Time, d time.Duration) *time.Time {
	n := t.Add(d)
	return &n
}

// seedExploreTopics creates the 6 default explore topics
func seedExploreTopics(_ *gorm.DB, now time.Time) []engmodel.ExploreTopic {
	now = now.Add(-30 * 24 * time.Hour) // Created 30 days ago
	return []engmodel.ExploreTopic{
		{
			Code:        "adventure",
			Name:        "Adventure Trips",
			NameAr:      "رحلات مغامرة",
			Description: "Thrilling experiences for adrenaline seekers - hiking, diving, safari and more",
			ImageURL:    "https://img.magnific.com/free-psd/travel-tourism-instagram-post-social-media-post-template_106176-2392.jpg",
			IconURL:     "https://img.magnific.com/free-psd/hiking-adventure-icon-template_47987-12345.jpg",
			SortOrder:   1,
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			Code:        "luxury",
			Name:        "Luxury Escapes",
			NameAr:      "هروب فاخر",
			Description: "Premium getaways with 5-star amenities, private villas, and exclusive experiences",
			ImageURL:    "https://d1csarkz8obe9u.cloudfront.net/posterpreviews/travel-agency-ad-design-template-b7db8661374e0eb66fad2ff933f59835_screen.jpg?ts=1653886566",
			IconURL:     "https://img.magnific.com/free-psd/luxury-travel-icon-template_47987-67890.jpg",
			SortOrder:   2,
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			Code:        "cultural",
			Name:        "Cultural Journeys",
			NameAr:      "رحلات ثقافية",
			Description: "Immersive heritage and culture tours - UNESCO sites, local traditions, history",
			ImageURL:    "https://img.magnific.com/free-psd/tour-travel-agency-instagram-post-social-media-post-web-banner-template-design_84443-25291.jpg?semt=ais_hybrid&w=740&q=80",
			IconURL:     "https://img.magnific.com/free-psd/cultural-travel-icon-template_47987-11111.jpg",
			SortOrder:   3,
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			Code:        "honeymoon",
			Name:        "Honeymoon Deals",
			NameAr:      "عروض شهر العسل",
			Description: "Romantic packages for newlyweds - private dinners, couples spa, sunset cruises",
			ImageURL:    "https://img.magnific.com/free-psd/travel-tourism-social-media-post-template_47987-29099.jpg?semt=ais_hybrid&w=740&q=80",
			IconURL:     "https://img.magnific.com/free-psd/honeymoon-icon-template_47987-22222.jpg",
			SortOrder:   4,
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			Code:        "family",
			Name:        "Family Holidays",
			NameAr:      "عطلات عائلية",
			Description: "Kid-friendly vacations for all ages - resorts with kids clubs, family activities",
			ImageURL:    "https://www.shutterstock.com/image-vector/travel-promo-vector-design-tour-260nw-2172252889.jpg",
			IconURL:     "https://img.magnific.com/free-psd/family-travel-icon-template_47987-33333.jpg",
			SortOrder:   5,
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			Code:        "friends",
			Name:        "Friends Getaways",
			NameAr:      "رحلات الأصدقاء",
			Description: "Group trips perfect for friends - adventure, nightlife, shared experiences",
			ImageURL:    "https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcTXRrGOt_cVkV4rlhQMvNUSNXTXDUZq0m1PWPuaeOa3rIyR-ub7nHtZZ4E&s=10",
			IconURL:     "https://img.magnific.com/free-psd/friends-travel-icon-template_47987-44444.jpg",
			SortOrder:   6,
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}
}

// seedPackageTopics associates packages with explore topics
func seedPackageTopics(pIstanbul, pDubai, pBali, pMaldives uint) []engmodel.PackageTopic {
	// We need topic IDs - we'll look them up by code
	// For seeding, we'll use the known IDs (1-6 based on sort order)
	// Istanbul (Turkey) -> Adventure, Cultural
	// Dubai (UAE) -> Luxury, Friends
	// Bali (Indonesia) -> Adventure, Cultural, Honeymoon
	// Maldives -> Luxury, Honeymoon, Family
	now := time.Now().Add(-24 * time.Hour)
	return []engmodel.PackageTopic{
		// Istanbul packages
		{PackageID: pIstanbul, TopicID: 1, AddedBy: 1, AddedAt: now}, // Adventure
		{PackageID: pIstanbul, TopicID: 3, AddedBy: 1, AddedAt: now}, // Cultural

		// Dubai packages
		{PackageID: pDubai, TopicID: 2, AddedBy: 1, AddedAt: now}, // Luxury
		{PackageID: pDubai, TopicID: 6, AddedBy: 1, AddedAt: now}, // Friends

		// Bali packages
		{PackageID: pBali, TopicID: 1, AddedBy: 1, AddedAt: now}, // Adventure
		{PackageID: pBali, TopicID: 3, AddedBy: 1, AddedAt: now}, // Cultural
		{PackageID: pBali, TopicID: 4, AddedBy: 1, AddedAt: now}, // Honeymoon

		// Maldives packages
		{PackageID: pMaldives, TopicID: 2, AddedBy: 1, AddedAt: now}, // Luxury
		{PackageID: pMaldives, TopicID: 4, AddedBy: 1, AddedAt: now}, // Honeymoon
		{PackageID: pMaldives, TopicID: 5, AddedBy: 1, AddedAt: now}, // Family
	}
}
