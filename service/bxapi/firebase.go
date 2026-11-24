package bxapi

import (
	"context"
	"errors"

	"firebase.google.com/go/v4/auth"
	"github.com/mo7amed4522/tayara_api/constants"
	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/model/bxapi"
	bxapiReq "github.com/mo7amed4522/tayara_api/model/bxapi/request"
	bxapiResp "github.com/mo7amed4522/tayara_api/model/bxapi/response"
	engageReq "github.com/mo7amed4522/tayara_api/model/engage/request"
	engageService "github.com/mo7amed4522/tayara_api/service/engage"
	"github.com/mo7amed4522/tayara_api/utils"
	fbutil "github.com/mo7amed4522/tayara_api/utils/firebase"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// FirebaseLogin authenticates an end-user via a Firebase ID token (Google / Apple / Phone). It
// verifies the token server-side, ensures the provider is enabled, then finds the local user by
// firebase_uid (creating one on first sign-in) and issues the Bx JWT.
func (s *UserServices) FirebaseLogin(ctx context.Context, req bxapiReq.FirebaseLoginReq) (*bxapiResp.UserLoginResponse, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	if !fbutil.Enabled() {
		return nil, errors.New("firebase auth is not enabled")
	}
	token, err := fbutil.VerifyIDToken(ctx, req.IDToken)
	if err != nil {
		global.GVA_LOG.Warn("firebase token verification failed", zap.Error(err))
		return nil, errors.New("invalid firebase id token")
	}
	provider := fbutil.SignInProvider(token)
	if err := fbutil.ValidateProvider(token, provider); err != nil {
		return nil, err
	}
	if err := validateFirebaseDevice(req); err != nil {
		return nil, err
	}
	// phone sign-ins expose the verified number in token.PhoneNumber
	var user bxapi.BxsUserProfile
	created := false
	err = global.GVA_DB.Where("firebase_uid = ?", token.UID).First(&user).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		// first sign-in — create the local account from the verified provider data
		user = buildUserFromFirebase(token, provider)
		if err := global.GVA_DB.Create(&user).Error; err != nil {
			return nil, err
		}
		created = true
	}
	if user.Status != constants.UserStatusActive {
		return nil, errors.New("account is blocked")
	}
	// issue the Bx JWT (X-token) for the mobile app
	jwtToken, claims, err := utils.LoginTokenWithAuthority(&user, constants.AuthorityUser)
	if err != nil {
		return nil, err
	}
	if created {
		if req.PushToken != "" {
			if err := (engageService.UserEngageServices{}).UpsertPushToken(user.ID, engageReq.PushTokenReq{
				DeviceID: req.DeviceID,
				Platform: req.Platform,
				Token:    req.PushToken,
			}); err != nil {
				global.GVA_LOG.Error("failed to register device token during Firebase signup", zap.Error(err))
			}
		}
		(engageService.NotificationServices{}).EmitWithImage(
			constants.NotifEventWelcome,
			constants.NotifRecipientUser,
			user.ID,
			constants.NotifChannelPush,
			"Welcome at Taiyara Marketplace",
			"We here to help you in any trip and booking you do as you wish",
			"https://raw.githubusercontent.com/tayara_app/refs/heads/master/assets/images/app_logo.png",
		)
	}
	return &bxapiResp.UserLoginResponse{
		User:      user,
		Token:     jwtToken,
		ExpiresAt: claims.RegisteredClaims.ExpiresAt.Unix() * 1000,
	}, nil
}

func validateFirebaseDevice(req bxapiReq.FirebaseLoginReq) error {
	hasDeviceDetails := req.DeviceID != "" || req.Platform != "" || req.PushToken != ""
	if !hasDeviceDetails {
		return nil
	}
	if req.DeviceID == "" || req.Platform == "" || req.PushToken == "" {
		return errors.New("device_id, platform, and push_token must be provided together")
	}
	if req.Platform != "ios" && req.Platform != "android" {
		return errors.New("platform must be ios or android")
	}
	return nil
}

// buildUserFromFirebase maps a verified Firebase identity onto a local user profile. Email, name,
// and phone are carried in the ID-token Claims map (the reserved provider claims); phone identity
// is also echoed in Firebase.Identities.phone.
func buildUserFromFirebase(token *auth.Token, provider string) bxapi.BxsUserProfile {
	user := bxapi.BxsUserProfile{
		FirebaseUID: token.UID,
		KYCStatus:   constants.KYCStatusNotSubmitted,
		Status:      constants.UserStatusActive,
		NickName:    "User",
	}
	claimStr := func(k string) string {
		if v, ok := token.Claims[k].(string); ok {
			return v
		}
		return ""
	}
	user.Email = claimStr("email")
	if name := claimStr("name"); name != "" {
		user.FullName = name
		user.NickName = name
	}
	switch provider {
	case fbutil.SignInProviderPhone:
		user.Phone = claimStr("phone_number")
		user.Username = "phone_" + token.UID
	default:
		if user.Email != "" {
			user.Username = user.Email
		} else {
			user.Username = "fb_" + token.UID
		}
	}
	return user
}
