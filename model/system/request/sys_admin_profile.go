package request

type SysRegisterReq struct {
	Username         string `json:"userName" form:"userName"`
	Password         string `json:"password" form:"password"`
	NickName         string `json:"nickName" form:"nickName"`
	HeaderImg        string `json:"headerImg" form:"headerImg"`
	Email            string `json:"email" form:"email"`
	PhoneCountryCode string `json:"phoneCountryCode" form:"phoneCountryCode"`
	Phone            string `json:"phone" form:"phone"`
	Department       string `json:"department" form:"department"`
	JobTitle         string `json:"jobTitle" form:"jobTitle"`
}

type SysLoginReq struct {
	Email    string `json:"email" form:"email"`
	Password string `json:"password" form:"password"`
}

type SysProfileUpdateReq struct {
	Username         string `json:"userName" form:"userName"`
	NickName         string `json:"nickName" form:"nickName"`
	HeaderImg        string `json:"headerImg" form:"headerImg"`
	Email            string `json:"email" form:"email"`
	PhoneCountryCode string `json:"phoneCountryCode" form:"phoneCountryCode"`
	Phone            string `json:"phone" form:"phone"`
	Department       string `json:"department" form:"department"`
	JobTitle         string `json:"jobTitle" form:"jobTitle"`
}

type SysChangePasswordReq struct {
	ID          uint   `json:"id" form:"id"`
	OldPassword string `json:"oldPassword" form:"oldPassword"`
	NewPassword string `json:"newPassword" form:"newPassword"`
}

type SysForgotPasswordReq struct {
	Email string `json:"email" form:"email"`
}

type SysResetPasswordReq struct {
	Token       string `json:"token" form:"token"`
	NewPassword string `json:"newPassword" form:"newPassword"`
}
