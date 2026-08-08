package utils

import (
	"errors"
	"time"

	jwt "github.com/golang-jwt/jwt/v4"

	"github.com/mo7amed4522/tayara_api/global"
	bxapiReq "github.com/mo7amed4522/tayara_api/model/bxapi/request"
)

type BxJWT struct {
	SigningKey []byte
}

var (
	BxTokenExpired     = errors.New("BxToken is expired")
	BxTokenNotValidYet = errors.New("BxToken not active yet")
	BxTokenMalformed   = errors.New("That's not even a bxToken")
	BxTokenInvalid     = errors.New("Couldn't handle this bxToken:")
)

func NewBxJWT() *BxJWT {
	return &BxJWT{
		[]byte(global.GVA_CONFIG.JWT.SigningKey),
	}
}

func (j *BxJWT) CreateClaims(baseClaims bxapiReq.BaseClaims) bxapiReq.CustomClaims {
	expiresDuration, _ := ParseDuration(global.GVA_CONFIG.JWT.ExpiresTime)
	claims := bxapiReq.CustomClaims{
		BaseClaims: baseClaims,
		BufferTime: 60 * 60 * 24 * 7, // 7-day buffer period
		RegisteredClaims: jwt.RegisteredClaims{
			Audience:  jwt.ClaimStrings{"BxAPI"},                           // Audience
			NotBefore: jwt.NewNumericDate(time.Now().Add(-1000)),           // Signature effective time
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresDuration)), // Expiration time
			Issuer:    global.GVA_CONFIG.JWT.Issuer,                        // Signature issuer
		},
	}
	return claims
}

// Create a token
func (j *BxJWT) CreateToken(claims bxapiReq.CustomClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.SigningKey)
}

// CreateTokenByOldToken Old token to new token using merge back source to avoid concurrency issues
func (j *BxJWT) CreateTokenByOldToken(oldToken string, claims bxapiReq.CustomClaims) (string, error) {
	v, err, _ := global.GVA_Concurrency_Control.Do("BxJWT:"+oldToken, func() (interface{}, error) {
		return j.CreateToken(claims)
	})
	return v.(string), err
}

// Parse token
func (j *BxJWT) ParseToken(tokenString string) (*bxapiReq.CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &bxapiReq.CustomClaims{}, func(token *jwt.Token) (i interface{}, e error) {
		// Pin the algorithm to HS256 to prevent algorithm-confusion forgery.
		if token.Method != jwt.SigningMethodHS256 {
			return nil, BxTokenInvalid
		}
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
