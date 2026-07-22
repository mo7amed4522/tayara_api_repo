package middleware

import (
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/utils"

	"github.com/gin-gonic/gin"
	"github.com/mo7amed4522/tayara_api/model/common/response"
)

func BxJWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token
		token := utils.GetToken(c)
		if token == "" {
			response.NoAuth("Not logged in or illegal access", c)
			c.Abort()
			return
		}

		// Use BxJWT to parse token
		j := utils.NewBxJWT()
		claims, err := j.ParseToken(token)
		if err != nil {
			if errors.Is(err, utils.BxTokenExpired) {
				response.NoAuth("Authorization has expired", c)
				utils.ClearToken(c)
				c.Abort()
				return
			}
			response.NoAuth(err.Error(), c)
			utils.ClearToken(c)
			c.Abort()
			return
		}

		// Set claims to context, using "bxClaims" as key
		c.Set("bxClaims", claims)

		// Directly extract user ID and set it to context, avoid duplicate parsing
		c.Set("bxUserID", claims.BaseClaims.ID)

		// Check if the token is about to expire, if so, refresh it
		if claims.ExpiresAt.Unix()-time.Now().Unix() < claims.BufferTime {
			dr, _ := utils.ParseDuration(global.GVA_CONFIG.JWT.ExpiresTime)
			claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(dr))
			newToken, _ := j.CreateTokenByOldToken(token, *claims)
			newClaims, _ := j.ParseToken(newToken)
			c.Header("new-token", newToken)
			c.Header("new-expires-at", strconv.FormatInt(newClaims.ExpiresAt.Unix(), 10))
			utils.SetToken(c, newToken, int(dr.Seconds()))
		}
		c.Next()

		// Set new token to response header
		if newToken, exists := c.Get("new-token"); exists {
			c.Header("new-token", newToken.(string))
		}
		if newExpiresAt, exists := c.Get("new-expires-at"); exists {
			c.Header("new-expires-at", newExpiresAt.(string))
		}
	}
}
