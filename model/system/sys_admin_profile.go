package system

import (
	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/model/common"
)

type Login interface {
	GetUsername() string
	GetNickname() string
	GetUserId() uint
	GetUserInfo() any
}

var _ Login = new(SysAdminProfile)

type SysAdminProfile struct {
	ID               uint           `json:"id" gorm:"primarykey;comment:id of admin"`                                                                  // UUID
	Username         string         `json:"userName" gorm:"index;comment:User login name"`                                                             // User login name
	Password         string         `json:"-"  gorm:"comment:Password"`                                                                                // Password
	NickName         string         `json:"nickName" gorm:"default:System User;comment:NickName"`                                                      // NickName
	HeaderImg        string         `json:"headerImg" gorm:"default:https://raw.githubusercontent.com/mo7amed4522/new_app/refs/heads/master/logo.png"` // HeaderImg
	PhoneCountryCode string         `json:"phoneCountryCode" gorm:"comment:User phone number country code"`                                            // User phone number country code
	Phone            string         `json:"phone"  gorm:"comment:Phone"`                                                                               // Phone
	Email            string         `json:"email"  gorm:"comment:Email"`                                                                               // Email
	Department       string         `json:"department" form:"department" gorm:"size:128;comment:Department"`
	JobTitle         string         `json:"jobTitle" form:"jobTitle" gorm:"size:128;comment:Job title"`
	Permissions      common.JSONMap `json:"permissions" form:"permissions" gorm:"type:json;comment:Extra permissions"`

	global.GVA_MODEL
}

func (SysAdminProfile) TableName() string {
	return "sys_admin_profile"
}

func (s *SysAdminProfile) GetUsername() string {
	return s.Username
}

func (s *SysAdminProfile) GetNickname() string {
	return s.NickName
}

func (s *SysAdminProfile) GetUserId() uint {
	return s.ID
}

func (s *SysAdminProfile) GetUserInfo() any {
	return *s
}
