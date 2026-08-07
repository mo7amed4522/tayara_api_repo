package bxapi

import (
	"errors"
	"time"

	"github.com/mo7amed4522/tayara_api/constants"
	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/model/bxapi"
	bxapiReq "github.com/mo7amed4522/tayara_api/model/bxapi/request"
	bxapiResp "github.com/mo7amed4522/tayara_api/model/bxapi/response"
	"github.com/mo7amed4522/tayara_api/utils"
)

type UserServices struct{}

// Register creates a new end-user account (public).
func (s *UserServices) Register(req bxapiReq.UserRegisterReq) (*bxapi.BxsUserProfile, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	var exists bxapi.BxsUserProfile
	if err := global.GVA_DB.Where("username = ?", req.Username).First(&exists).Error; err == nil {
		return nil, errors.New("username already registered")
	}
	if err := global.GVA_DB.Where("email = ?", req.Email).First(&exists).Error; err == nil {
		return nil, errors.New("email already registered")
	}
	nick := req.NickName
	if nick == "" {
		nick = req.Username
	}
	user := &bxapi.BxsUserProfile{
		FullName:         req.FullName,
		Username:         req.Username,
		Password:         utils.BcryptHash(req.Password),
		NickName:         nick,
		Email:            req.Email,
		PhoneCountryCode: req.PhoneCountryCode,
		Phone:            req.Phone,
		Nationality:      req.Nationality,
		Status:           constants.UserStatusActive,
	}
	if err := global.GVA_DB.Create(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

// Login authenticates a user and issues the Bx (X-token) JWT.
func (s *UserServices) Login(req bxapiReq.UserLoginReq) (*bxapiResp.UserLoginResponse, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	var user bxapi.BxsUserProfile
	if err := global.GVA_DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		return nil, errors.New("invalid username or password")
	}
	if !utils.BcryptCheck(req.Password, user.Password) {
		return nil, errors.New("invalid username or password")
	}
	if user.Status != constants.UserStatusActive {
		return nil, errors.New("account is blocked")
	}
	token, claims, err := utils.LoginTokenWithAuthority(&user, constants.AuthorityUser)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	_ = global.GVA_DB.Model(&bxapi.BxsUserProfile{}).Where("id = ?", user.ID).
		Update("last_login_at", &now).Error
	return &bxapiResp.UserLoginResponse{
		User:      user,
		Token:     token,
		ExpiresAt: claims.RegisteredClaims.ExpiresAt.Unix() * 1000,
	}, nil
}

// GetProfile returns a user's own profile by id.
func (s *UserServices) GetProfile(userID uint) (*bxapi.BxsUserProfile, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	var user bxapi.BxsUserProfile
	if err := global.GVA_DB.First(&user, userID).Error; err != nil {
		return nil, errors.New("user not found")
	}
	return &user, nil
}

// ChangePassword updates the user's own password after verifying the old one.
func (s *UserServices) ChangePassword(userID uint, req bxapiReq.ChangeUserPasswordReq) error {
	if global.GVA_DB == nil {
		return errors.New("db not init")
	}
	var user bxapi.BxsUserProfile
	if err := global.GVA_DB.First(&user, userID).Error; err != nil {
		return errors.New("user not found")
	}
	if !utils.BcryptCheck(req.OldPassword, user.Password) {
		return errors.New("old password is incorrect")
	}
	return global.GVA_DB.Model(&bxapi.BxsUserProfile{}).Where("id = ?", userID).
		Update("password", utils.BcryptHash(req.NewPassword)).Error
}

// --- admin actions ---

// Block toggles a user's block/unblock status.
func (s *UserServices) Block(adminID uint, req bxapiReq.UserBlockReq) (*bxapi.BxsUserProfile, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	var user bxapi.BxsUserProfile
	if err := global.GVA_DB.First(&user, req.UserID).Error; err != nil {
		return nil, errors.New("user not found")
	}
	target := constants.UserStatusBlocked
	if user.Status == constants.UserStatusBlocked {
		target = constants.UserStatusActive
	}
	if err := global.GVA_DB.Model(&bxapi.BxsUserProfile{}).Where("id = ?", user.ID).
		Updates(map[string]interface{}{"status": target, "block_reason": req.Reason}).Error; err != nil {
		return nil, err
	}
	user.Status = target
	return &user, nil
}

// List returns users (admin: all; scoped by status filter).
func (s *UserServices) List(req bxapiReq.UserPageReq) ([]bxapi.BxsUserProfile, int64, error) {
	if global.GVA_DB == nil {
		return nil, 0, errors.New("db not init")
	}
	db := global.GVA_DB.Model(&bxapi.BxsUserProfile{})
	if req.Status != "" {
		db = db.Where("status = ?", req.Status)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []bxapi.BxsUserProfile
	if err := db.Scopes(req.Paginate()).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
