package utils

import (
	"errors"
	"time"

	jwt "github.com/golang-jwt/jwt/v4"

	"github.com/mo7amed4522/tayara_api/global"
	bxapiReq "github.com/mo7amed4522/tayara_api/model/bxapi/request"
	dxapiReq "github.com/mo7amed4522/tayara_api/model/dxapi/request"
)

type DxJWT struct {
	SigningKey []byte
}

var (
	DxTokenExpired     = errors.New("DxToken is expired")
	DxTokenNotValidYet = errors.New("DxToken not active yet")
	DxTokenMalformed   = errors.New("That's not even a dxToken")
	DxTokenInvalid     = errors.New("Couldn't handle this dxToken:")
)

func NewDxJWT() *DxJWT {
	return &DxJWT{
		[]byte(global.GVA_CONFIG.JWT.SigningKey),
	}
}

func (j *DxJWT) CreateDxClaims(baseClaims dxapiReq.BaseClaims) dxapiReq.CustomClaims {
	expiresDuration, _ := ParseDuration(global.GVA_CONFIG.JWT.ExpiresTime)
	claims := dxapiReq.CustomClaims{
		BaseClaims: baseClaims,
		BufferTime: 60 * 60 * 24 * 7, // 7-day buffer period
		RegisteredClaims: jwt.RegisteredClaims{
			Audience:  jwt.ClaimStrings{"DxAPI"},                           // Audience
			NotBefore: jwt.NewNumericDate(time.Now().Add(-1000)),           // Signature effective time
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresDuration)), // Expiration time
			Issuer:    global.GVA_CONFIG.JWT.Issuer,                        // Signature issuer
		},
	}
	return claims
}

// Create a token
func (j *DxJWT) CreateDxToken(claims bxapiReq.CustomClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.SigningKey)
}

// CreateTokenByOldToken Old token to new token using merge back source to avoid concurrency issues
func (j *DxJWT) CreateTokenByOldToken(oldToken string, claims bxapiReq.CustomClaims) (string, error) {
	v, err, _ := global.GVA_Concurrency_Control.Do("DxJWT:"+oldToken, func() (interface{}, error) {
		return j.CreateDxToken(claims)
	})
	return v.(string), err
}

// Parse token
func (j *DxJWT) ParseDxToken(tokenString string) (*bxapiReq.CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &bxapiReq.CustomClaims{}, func(token *jwt.Token) (i interface{}, e error) {
		return j.SigningKey, nil
	})
	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				return nil, BxTokenMalformed
			} else if ve.Errors&jwt.ValidationErrorExpired != 0 {
				// The token has expired
				return nil, BxTokenExpired
			} else if ve.Errors&jwt.ValidationErrorNotValidYet != 0 {
				return nil, BxTokenNotValidYet
			} else {
				return nil, BxTokenInvalid
			}
		}
	}
	if token != nil {
		if claims, ok := token.Claims.(*bxapiReq.CustomClaims); ok && token.Valid {
			return claims, nil
		}
		return nil, BxTokenInvalid

	} else {
		return nil, BxTokenInvalid
	}
}
