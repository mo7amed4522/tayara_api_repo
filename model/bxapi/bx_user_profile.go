package bxapi

import (
	"time"

	"github.com/mo7amed4522/tayara_api/global"
)

type Login interface {
	GetUsername() string
	GetNickname() string
	GetUserId() uint
	GetUserInfo() any
}

var _ Login = new(BxsUserProfile)

type BxsUserProfile struct {
	ID               uint       `json:"id" gorm:"primarykey;comment:id of User"`
	FullName         string     `json:"fullName" form:"fullName" gorm:"size:191;index;comment:Full name"`
	Username         string     `json:"userName" gorm:"index;comment:User login name"`                                                             // User login name
	Password         string     `json:"-"  gorm:"comment:Password"`                                                                                // Password
	NickName         string     `json:"nickName" gorm:"default:System User;comment:NickName"`                                                      // NickName
	HeaderImg        string     `json:"headerImg" gorm:"default:https://raw.githubusercontent.com/mo7amed4522/new_app/refs/heads/master/logo.png"` // HeaderImg
	PhoneCountryCode string     `json:"phoneCountryCode" gorm:"comment:User phone number country code"`                                            // User phone number country code
	Phone            string     `json:"phone"  gorm:"comment:Phone"`                                                                               // Phone
	Email            string     `json:"email"  gorm:"comment:Email"`                                                                               // Email
	Nationality      string     `json:"nationality" form:"nationality" gorm:"size:128;index;comment:Nationality"`
	DateOfBirth      *time.Time `json:"dateOfBirth" form:"dateOfBirth" gorm:"comment:Date of birth"`
	KYCStatus        string     `json:"kycStatus" form:"kycStatus" gorm:"size:32;index;default:pending;comment:KYC status"`
	TotalSpend       int64      `json:"totalSpend" form:"totalSpend" gorm:"default:0;comment:Total spend in minor units"`
	Status           string     `json:"status" form:"status" gorm:"size:32;index;default:active;comment:Client status"`
	FirebaseUID      string     `json:"firebaseUid" gorm:"size:128;index;comment:Firebase auth UID"`                                // Firebase UID
	AuthProvider     string     `json:"authProvider" gorm:"size:32;index;comment:Auth provider: email,google,apple,facebook,phone"` // Auth provider
	global.GVA_MODEL
}

func (BxsUserProfile) TableName() string {
	return "bx_user_profile"
}

func (s *BxsUserProfile) GetUsername() string {
	return s.Username
}

func (s *BxsUserProfile) GetNickname() string {
	return s.NickName
}

func (s *BxsUserProfile) GetUserId() uint {
	return s.ID
}

func (s *BxsUserProfile) GetUserInfo() any {
	return *s
}
