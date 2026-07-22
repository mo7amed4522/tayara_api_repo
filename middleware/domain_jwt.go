package middleware

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mo7amed4522/tayara_api/model/common/response"
	"github.com/mo7amed4522/tayara_api/utils"
)

// domainJWTAuth handles JWT authentication for domain-specific token headers.
// For SystemTokenHeader (S-token) it also checks the token blacklist so that
// logged-out admin tokens are immediately invalidated.

const (
	BxTokenHeader     = "X-token"
	ManageTokenHeader = "A-token"
	SystemTokenHeader = "S-token"
)

func BxDomainJWTAuth() gin.HandlerFunc {
	return domainJWTAuth(BxTokenHeader, "bxClaims", "bxUserID")
}

func ManageJWTAuth() gin.HandlerFunc {
	return domainJWTAuth(ManageTokenHeader, "manageClaims", "manageUserID")
}

func SystemJWTAuth() gin.HandlerFunc {
	return domainJWTAuth(SystemTokenHeader, "claims", "systemUserID")
}

func domainJWTAuth(headerName, claimsKey, userIDKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader(headerName)
		if token == "" {
			token = c.GetHeader(strings.ToLower(headerName))
		}
		if token == "" {
			cookieToken, err := c.Cookie(headerName)
			if err == nil && cookieToken != "" {
				token = cookieToken
			}
		}
		if token == "" {
			cookieToken, err := c.Cookie(strings.ToLower(headerName))
			if err == nil && cookieToken != "" {
				token = cookieToken
			}
		}
		if token == "" {
			response.NoAuth("Not logged in or illegal access", c)
			c.Abort()
			return
		}
		// Check blacklist so logout is effective immediately for all token types.
		if jwtService.IsBlacklist(token) {
			response.NoAuth("Token has been invalidated. Please log in again.", c)
			c.Abort()
			return
		}
		j := utils.NewJWT()
		claims, err := j.ParseToken(token)
		if err != nil {
			if errors.Is(err, utils.TokenExpired) {
				response.NoAuth("Authorization has expired", c)
				c.Abort()
				return
			}
			response.NoAuth(err.Error(), c)
			c.Abort()
			return
		}
		c.Set(claimsKey, claims)
		c.Set(userIDKey, claims.BaseClaims.ID)
		c.Next()
	}
}
