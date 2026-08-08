package system

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"

	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/model/system"
	"github.com/mo7amed4522/tayara_api/model/system/request"
	"github.com/mo7amed4522/tayara_api/utils"
	"github.com/mo7amed4522/tayara_api/utils/upload"
	"gorm.io/gorm"
)

type SysAdminProfileServices struct{}

var jwtService = &JwtService{}

var SysAdminProfileServiceApp = new(SysAdminProfileServices)

// @author: [mo7amed4522](https://github.com/mo7amed4522)
// @function: Register
// @description: Register
// @param: u model.SysAdminProfile
// @return: admin system.SysAdminProfile, err error
func (sysAdminProfileServices *SysAdminProfileServices) Register(u request.SysRegisterReq, headerImage *multipart.FileHeader) (userInter system.SysAdminProfile, err error) {
	var user system.SysAdminProfile
	if !errors.Is(global.GVA_DB.Where("email = ?", u.Email).First(&user).Error, gorm.ErrRecordNotFound) { // Check if the username is registered
		return userInter, errors.New("email already registered")
	}
	if headerImage != nil {
		headerImgURL, _, err := upload.NewOss().UploadFileWithPath(headerImage, "system/profile-images")
		if err != nil {
			return system.SysAdminProfile{}, err
		}
		u.HeaderImg = headerImgURL
	}
	user = system.SysAdminProfile{
		Username:         u.Username,
		Password:         utils.BcryptHash(u.Password),
		NickName:         u.NickName,
		PhoneCountryCode: u.PhoneCountryCode,
		Phone:            u.Phone,
		Email:            u.Email,
		HeaderImg:        u.HeaderImg,
		Department:       u.Department,
		JobTitle:         u.JobTitle,
	}
	if err := global.GVA_DB.Create(&user).Error; err != nil {
		return user, err
	}
	return
}

// @author: [mo7amed4522](https://github.com/mo7amed4522)
// @function: Login
// @description: Login
// @param: u *model.SysAdminProfile
// @return: err error, adminInter *model.SysAdminProfile
// dummyPasswordHash is a valid bcrypt hash used only to keep login timing uniform
// when the email does not exist, so attackers cannot enumerate accounts by timing.
var dummyPasswordHash = utils.BcryptHash("invalid-account-timing")

func (sysAdminProfileServices *SysAdminProfileServices) Login(u request.SysLoginReq) (userInter *system.SysAdminProfile, err error) {
	if nil == global.GVA_DB {
		return nil, fmt.Errorf("db not init")
	}
	var user system.SysAdminProfile
	err = global.GVA_DB.Where("email = ?", u.Email).First(&user).Error
	if err == nil {
		if ok := utils.BcryptCheck(u.Password, user.Password); !ok {
			return nil, errors.New("invalid email or password")
		}
		return &user, nil
	}
	// Account not found: burn a bcrypt comparison so response timing does not reveal
	// whether the email is registered, then return the same generic error. This
	// removes the previous "record not found" vs "password error" enumeration leak.
	utils.BcryptCheck(u.Password, dummyPasswordHash)
	return nil, errors.New("invalid email or password")
}

// @author: [mo7amed4522](https://github.com/mo7amed4522)
// @function: ChangePassword
// @description: Change password
// @param: u *model.SysAdminProfile
// @return: userInter *model.SysAdminProfile,err error
func (sysAdminProfileServices *SysAdminProfileServices) ChangePassword(u *request.SysChangePasswordReq) (err error) {
	var user system.SysAdminProfile
	if err = global.GVA_DB.Where("id = ?", u.ID).First(&user).Error; err != nil {
		return err
	}
	if ok := utils.BcryptCheck(u.OldPassword, user.Password); !ok {
		return errors.New("original password error")
	}
	user.Password = utils.BcryptHash(u.NewPassword)
	err = global.GVA_DB.Save(&user).Error
	return err

}

// @author: [mo7amed4522](https://github.com/mo7amed4522)
// @function: UpdatePassword
// @description: Update password
// @param: u *model.SysAdminProfile, newPassword string
// @return: userInter *model.SysAdminProfile,err error
func (sysAdminProfileServices *SysAdminProfileServices) UpdatePassword(u *system.SysAdminProfile, newPassword string) (userInter *system.SysAdminProfile, err error) {
	var user system.SysAdminProfile
	if err = global.GVA_DB.Where("id = ?", u.ID).First(&user).Error; err != nil {
		return nil, err
	}
	user.Password = utils.BcryptHash(newPassword)
	err = global.GVA_DB.Save(&user).Error
	return &user, err

}

// @author: [mo7amed4522](https://github.com/mo7amed4522)
// @function: SetSelfInfo
// @description: Set user info
// @param: reqUser model.SysAdminProfile
// @return: err error, user system.SysAdminProfile
func (sysAdminProfileServices *SysAdminProfileServices) SetSelfInfo(req system.SysAdminProfile) error {
	return global.GVA_DB.Model(&system.SysAdminProfile{}).
		Where("id=?", req.ID).
		Updates(req).Error
}

// @author: [mo7amed4522](https://github.com/mo7amed4522)
// @function: FindUserById
// @description: Get user info by id
// @param: id int
// @return: err error, user *system.SysAdminProfile
func (sysAdminProfileServices *SysAdminProfileServices) FindUserById(id int) (user *system.SysAdminProfile, err error) {
	var u system.SysAdminProfile
	err = global.GVA_DB.Where("id = ?", id).First(&u).Error
	return &u, err
}

// @author: [mo7amed4522](https://github.com/mo7amed4522)
// @function: Logout
// @description: Logout user
// @param: username string
// @return: err error
func (sysAdminProfileServices *SysAdminProfileServices) Logout(username string) error {
	if username == "" {
		return errors.New("username cannot be empty")
	}
	jwtStr, err := jwtService.GetRedisJWT(username)
	if err != nil {
		return nil
	}

	var blackJWT system.JwtBlacklist
	blackJWT.Jwt = jwtStr
	if err := jwtService.JsonInBlacklist(blackJWT); err != nil {
		return fmt.Errorf("failed to add JWT to blacklist: %v", err)
	}

	if err := global.GVA_REDIS.Del(context.Background(), username).Err(); err != nil {
		return fmt.Errorf("failed to delete JWT from Redis: %v", err)
	}

	return nil
}
