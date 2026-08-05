package request

import (
	commonReq "github.com/mo7amed4522/tayara_api/model/common/request"
)

// UserRegisterReq public registration payload for an end-user.
type UserRegisterReq struct {
	FullName         string `json:"fullName" binding:"required"`
	Username         string `json:"username" binding:"required"`
	Password         string `json:"password" binding:"required,min=6"`
	NickName         string `json:"nickName"`
	Email            string `json:"email" binding:"required,email"`
	PhoneCountryCode string `json:"phoneCountryCode"`
	Phone            string `json:"phone"`
	Nationality      string `json:"nationality"`
}

// UserLoginReq user login credentials.
type UserLoginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// UserBlockReq admin blocks/unblocks a user.
type UserBlockReq struct {
	UserID uint   `json:"userId" binding:"required"`
	Reason string `json:"reason"`
}

// UserPageReq paginated user listing (admin) with optional status filter.
type UserPageReq struct {
	commonReq.PageInfo
	Status string `json:"status" form:"status"`
}

// ChangeUserPasswordReq user changes their own password.
type ChangeUserPasswordReq struct {
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required,min=6"`
}
