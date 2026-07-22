package utils

import (
	"errors"
	"time"

	jwt "github.com/golang-jwt/jwt/v4"

	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/model/system/request"
)

type JWT struct {
	SigningKey []byte
}

var (
	TokenExpired     = errors.New("Token is expired")
	TokenNotValidYet = errors.New("Token not active yet")
	TokenMalformed   = errors.New("That's not even a token")
	TokenInvalid     = errors.New("Couldn't handle this token:")
)

func NewJWT() *JWT {
	return &JWT{
		[]byte(global.GVA_CONFIG.JWT.SigningKey),
	}
}

func (j *JWT) CreateClaims(baseClaims request.BaseClaims) request.CustomClaims {
	bf, _ := ParseDuration(global.GVA_CONFIG.JWT.BufferTime)
	ep, _ := ParseDuration(global.GVA_CONFIG.JWT.ExpiresTime)
	claims := request.CustomClaims{
		BaseClaims: baseClaims,
		BufferTime: int64(bf / time.Second), // The buffer period is 1 day. During this period, a new token will be obtained to refresh the token. At this time, a user will have two valid tokens, but the front end will only keep one and the other will be lost.
		RegisteredClaims: jwt.RegisteredClaims{
			Audience:  jwt.ClaimStrings{"GVA"},                   // Audience
			NotBefore: jwt.NewNumericDate(time.Now().Add(-1000)), // Signature effective time
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ep)),    // Expiration time 7 days   configuration file
			Issuer:    global.GVA_CONFIG.JWT.Issuer,              // Signature issuer
		},
	}
	return claims
}

// Create a token
func (j *JWT) CreateToken(claims request.CustomClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.SigningKey)
}

// CreateTokenByOldToken Old token to new token using merge back source to avoid concurrency issues
func (j *JWT) CreateTokenByOldToken(oldToken string, claims request.CustomClaims) (string, error) {
	v, err, _ := global.GVA_Concurrency_Control.Do("JWT:"+oldToken, func() (interface{}, error) {
		return j.CreateToken(claims)
	})
	return v.(string), err
}

// Parse token
func (j *JWT) ParseToken(tokenString string) (*request.CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &request.CustomClaims{}, func(token *jwt.Token) (i interface{}, e error) {
		return j.SigningKey, nil
	})
	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				return nil, TokenMalformed
			} else if ve.Errors&jwt.ValidationErrorExpired != 0 {
				return nil, TokenExpired
			} else if ve.Errors&jwt.ValidationErrorNotValidYet != 0 {
				return nil, TokenNotValidYet
			} else {
				return nil, TokenInvalid
			}
		}
	}
	if token != nil {
		if claims, ok := token.Claims.(*request.CustomClaims); ok && token.Valid {
			return claims, nil
		}
		return nil, TokenInvalid

	} else {
		return nil, TokenInvalid
	}
}
