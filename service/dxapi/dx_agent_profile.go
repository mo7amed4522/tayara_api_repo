package dxapi

import (
	"errors"
	"fmt"
	"mime/multipart"

	"github.com/mo7amed4522/tayara_api/constants"
	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/model/dxapi"
	"github.com/mo7amed4522/tayara_api/model/dxapi/request"
	"github.com/mo7amed4522/tayara_api/model/dxapi/response"
	"github.com/mo7amed4522/tayara_api/utils"
	"github.com/mo7amed4522/tayara_api/utils/upload"
	"gorm.io/gorm"
)

type DxAgentProfileService struct{}

// @author: [mo7amed4522](https://github.com/mo7amed4522)
// @function: Register
// @description: Register
// @param: u dxapi.DxAgentProfile
// @return: agent dxapi.DxAgentProfile, err error
func (dxAgentProfileService DxAgentProfileService) Register(req request.DxCreateAgentProfileRequest, headerImage *multipart.FileHeader) (res dxapi.DxAgentProfile, err error) {
	var agent dxapi.DxAgentProfile
	if !errors.Is(global.GVA_DB.Where("business_email = ?", req.BusinessEmail).First(&agent).Error, gorm.ErrRecordNotFound) { // Check if the username is registered
		userInter := dxapi.DxAgentProfile{}
		return userInter, errors.New("email already registered")
	}
	if headerImage != nil {
		headerImgURL, _, err := upload.NewOss().UploadFileWithPath(headerImage, "system/profile-images")
		if err != nil {
			return dxapi.DxAgentProfile{}, err
		}
		req.HeaderImg = headerImgURL
	}
	agent = dxapi.DxAgentProfile{
		CompanyName:       req.CompanyName,
		Password:          utils.BcryptHash(req.Password),
		ResponsibleName:   req.ResponsibleName,
		PhoneCountryCode:  req.PhoneCountryCode,
		Phone:             req.Phone,
		BusinessEmail:     req.BusinessEmail,
		HeaderImg:         req.HeaderImg,
		Country:           req.Country,
		CompanyAddress:    req.CompanyAddress,
		LicenseNumber:     req.LicenseNumber,
		PlanID:            req.PlanID,
		ShortDescription:  req.ShortDescription,
		LongDescription:   req.LongDescription,
		CommissionPercent: req.CommissionPercent,
		SLAHours:          req.SLAHours,
	}
	if err := global.GVA_DB.Create(&agent).Error; err != nil {
		return agent, err
	}
	return
}

// @author: [mo7amed4522](https://github.com/mo7amed4522)
// @function: Login
// @description: Login
// @param: u *dxapi.DxAgentProfile
// @return: err error, res *dxapi.DxAgentProfile
func (dxAgentProfileService DxAgentProfileService) Login(req request.DxLoginAgentRequest) (res *response.DxAgentLoginResponse, err error) {
	if nil == global.GVA_DB {
		return nil, fmt.Errorf("db not init")
	}

	var user dxapi.DxAgentProfile
	err = global.GVA_DB.Where("business_email = ?", req.Email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid credentials")
		}
		return nil, err
	}

	// Check password first for security
	if !utils.BcryptCheck(req.Password, user.Password) {
		return nil, errors.New("invalid credentials")
	}

	// Handle different status cases
	switch user.Status {
	case constants.AgentStatusPending:
		return nil, errors.New("your account is pending approval. Please wait for admin verification")
	case constants.AgentStatusRejected:
		return nil, errors.New("your account has been rejected. Please contact support for more information")
	case constants.AgentStatusActive:
		return res, nil
	default:
		return nil, errors.New("invalid account status")
	}
}
