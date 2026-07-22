package middleware

import (
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/model/common/response"
	"github.com/mo7amed4522/tayara_api/model/system"
	"github.com/mo7amed4522/tayara_api/service"
	"github.com/mo7amed4522/tayara_api/utils"

	"github.com/gin-gonic/gin"
)

var jwtService = service.ServiceGroupApp.SystemServiceGroup.JwtService

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Here we use jwt authorization to get the header information x-token, the token information will be returned when logging in, the front end needs to store the token in the cookie or localStorage, but it needs to be negotiated with the backend about the expiration time, it can be agreed to refresh the token or re-login
		token := utils.GetToken(c)
		if token == "" {
			response.NoAuth("Not logged in or illegal access", c)
			c.Abort()
			return
		}
		if jwtService.IsBlacklist(token) {
			response.NoAuth("Your account is logged in from another location or the token has expired", c)
			utils.ClearToken(c)
			c.Abort()
			return
		}
		j := utils.NewJWT()
		// parseToken Parse the information contained in the token
		claims, err := j.ParseToken(token)
		if err != nil {
			if errors.Is(err, utils.TokenExpired) {
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

		// The logged-in user has been disabled by the administrator, need to invalidate the jwt of the user, this is relatively performance-intensive, if needed, please open it yourself
		// The logic of the user being deleted needs to be optimized, this is relatively performance-intensive, if needed, please open it yourself

		if user, err := userService.FindUserByID(claims.BaseClaims.ID); err != nil || user.Enable == 2 {
			_ = jwtService.JsonInBlacklist(system.JwtBlacklist{Jwt: token})
			response.FailWithDetailed(gin.H{"reload": true}, err.Error(), c)
			c.Abort()
		}
		c.Set("claims", claims)
		if claims.ExpiresAt.Unix()-time.Now().Unix() < claims.BufferTime {
			dr, _ := utils.ParseDuration(global.GVA_CONFIG.JWT.ExpiresTime)
			claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(dr))
			newToken, _ := j.CreateTokenByOldToken(token, *claims)
			newClaims, _ := j.ParseToken(newToken)
			c.Header("new-token", newToken)
			c.Header("new-expires-at", strconv.FormatInt(newClaims.ExpiresAt.Unix(), 10))
			utils.SetToken(c, newToken, int(dr.Seconds()))
			if global.GVA_CONFIG.System.UseMultipoint {
				// Record the new active jwt
				_ = jwtService.SetRedisJWT(newToken, newClaims.Username)
			}
		}
		c.Next()

		if newToken, exists := c.Get("new-token"); exists {
			c.Header("new-token", newToken.(string))
		}
		if newExpiresAt, exists := c.Get("new-expires-at"); exists {
			c.Header("new-expires-at", newExpiresAt.(string))
		}
	}
}
