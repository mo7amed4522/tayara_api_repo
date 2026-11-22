package booking

import (
	"time"

	bookingmodel "github.com/mo7amed4522/tayara_api/model/booking"
	"gorm.io/gorm"
)

// userID resolves a user ID by username or email.
func userID(db *gorm.DB, usernameOrEmail string) uint {
	var id uint
	db.Table("bx_user_profile").
		Select("id").
		Where("username = ? OR email = ?", usernameOrEmail, usernameOrEmail).
		Limit(1).Scan(&id)
	return id
}

// agentID resolves an agent ID by email.
func agentID(db *gorm.DB, email string) uint {
	var id uint
	db.Table("bx_agent_profile").
		Select("id").
		Where("email = ?", email).
		Limit(1).Scan(&id)
	return id
}

// packageID resolves a package ID by package_code.
func packageID(db *gorm.DB, code string) uint {
	var id uint
	db.Table("bx_package").
		Select("id").
		Where("package_code = ?", code).
		Limit(1).Scan(&id)
	return id
}

// countryID resolves a country ID by ISO alpha-2 code.
func countryID(db *gorm.DB, iso2 string) uint {
	var id uint
	db.Table("bx_country").
		Select("id").
		Where("iso2 = ?", iso2).
		Limit(1).Scan(&id)
	return id
}

// docTypeID resolves a document type ID by slug/code.
func docTypeID(db *gorm.DB, slug string) uint {
	var id uint
	db.Table("bx_document_type").
		Select("id").
		Where("slug = ?", slug).
		Limit(1).Scan(&id)
	return id
}

// bookingByCode resolves a booking ID by booking_code.
func bookingByCode(db *gorm.DB, code string) uint {
	var b bookingmodel.Booking
	db.Select("id").Where("booking_code = ?", code).First(&b)
	return b.ID
}

// tAdd returns now + hours*duration. Negative hours = past.
func tAdd(now time.Time, d time.Duration) *time.Time {
	t := now.Add(d)
	return &t
}

// uintPtr returns a pointer to a uint value.
func uintPtr(v uint) *uint {
	return &v
}
