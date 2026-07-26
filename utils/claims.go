package utils

import (
	"net"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mo7amed4522/tayara_api/global"
	bxapiReq "github.com/mo7amed4522/tayara_api/model/bxapi/request"
	"github.com/mo7amed4522/tayara_api/model/dxapi"
	dxapiReq "github.com/mo7amed4522/tayara_api/model/dxapi/request"
	"github.com/mo7amed4522/tayara_api/model/system"
	systemReq "github.com/mo7amed4522/tayara_api/model/system/request"
)

func ClearToken(c *gin.Context) {
	// Add cookie x-token to the source web
	domain := GetCookieDomain(c.Request.Host)
	c.SetCookie("x-token", "", -1, "/", domain, false, false)
}

func SetToken(c *gin.Context, token string, maxAge int) {
	// Add cookie x-token to the source web
	domain := GetCookieDomain(c.Request.Host)
	c.SetCookie("x-token", token, maxAge, "/", domain, false, false)
}

// GetCookieDomain returns a domain suitable for use in `http.SetCookie`.
// It strips any port and returns empty string for IPs (so cookie becomes host-only).
func GetCookieDomain(rawHost string) string {
	host := rawHost
	if h, _, err := net.SplitHostPort(rawHost); err == nil {
		host = h
	} else {
		// handle IPv6 without brackets like "::1:8888"
		if strings.Count(rawHost, ":") > 1 {
			if i := strings.LastIndex(rawHost, ":"); i > 0 {
				host = rawHost[:i]
			}
		}
	}
	if net.ParseIP(host) != nil {
		return ""
	}
	return host
}

func GetToken(c *gin.Context) string {
	token, _ := c.Cookie("x-token")
	if token == "" {
		j := NewJWT()
		token = c.Request.Header.Get("x-token")
		claims, err := j.ParseToken(token)
		if err != nil {
			global.GVA_LOG.Error("Failed to re-write cookie token, failed to parse token, please check if the request header contains x-token and claims is in the correct structure")
			return token
		}
		SetToken(c, token, int((claims.ExpiresAt.Unix()-time.Now().Unix())/60))
	}
	return token
}

func GetClaims(c *gin.Context) (*systemReq.CustomClaims, error) {
	token := GetToken(c)
	j := NewJWT()
	claims, err := j.ParseToken(token)
	if err != nil {
		global.GVA_LOG.Error("Failed to get information from Gin's Context after parsing with jwt, please check if the request header contains x-token and claims is in the correct structure")
	}
	return claims, err
}

// GetUserID Get the user ID from the Gin's Context after parsing with jwt
func GetUserID(c *gin.Context) uint {
	if claims, exists := c.Get("claims"); !exists {
		if cl, err := GetClaims(c); err != nil {
			return 0
		} else {
			return cl.BaseClaims.ID
		}
	} else {
		waitUse := claims.(*systemReq.CustomClaims)
		return waitUse.BaseClaims.ID
	}
}

// GetUserAuthorityId Get the user role id from the Gin's Context after parsing with jwt
func GetUserAuthorityId(c *gin.Context) uint {
	if claims, exists := c.Get("claims"); !exists {
		if cl, err := GetClaims(c); err != nil {
			return 0
		} else {
			return cl.AuthorityId
		}
	} else {
		waitUse := claims.(*systemReq.CustomClaims)
		return waitUse.AuthorityId
	}
}

// GetUserInfo Get the user role id from the Gin's Context after parsing with jwt
func GetUserInfo(c *gin.Context) *systemReq.CustomClaims {
	if claims, exists := c.Get("claims"); !exists {
		if cl, err := GetClaims(c); err != nil {
			return nil
		} else {
			return cl
		}
	} else {
		waitUse := claims.(*systemReq.CustomClaims)
		return waitUse
	}
}

// GetUserName Get the username from the Gin's Context after parsing with jwt
func GetUserName(c *gin.Context) string {
	if claims, exists := c.Get("claims"); !exists {
		if cl, err := GetClaims(c); err != nil {
			return ""
		} else {
			return cl.Username
		}
	} else {
		waitUse := claims.(*systemReq.CustomClaims)
		return waitUse.Username
	}
}

func LoginToken(user system.Login) (token string, claims systemReq.CustomClaims, err error) {
	j := &JWT{SigningKey: []byte(global.GVA_CONFIG.JWT.SigningKey)} // Unique signature
	claims = j.CreateClaims(systemReq.BaseClaims{
		ID:       user.GetUserId(),
		NickName: user.GetNickname(),
		Username: user.GetUsername(),
	})
	token, err = j.CreateToken(claims)
	if err != nil {
		return
	}
	return
}

// GetBxClaims Get the bxapi claims from the Gin's Context after parsing with jwt
func GetBxClaims(c *gin.Context) (*bxapiReq.CustomClaims, error) {
	token := GetToken(c)
	j := NewBxJWT()
	claims, err := j.ParseToken(token)
	if err != nil {
		global.GVA_LOG.Error("Failed to get information from Gin's Context after parsing with jwt, please check if the request header contains x-token and claims is in the correct structure")
	}
	return claims, err
}

// GetBxUserID Get the bxapi user ID from the Gin's Context after parsing with jwt
func GetBxUserID(c *gin.Context) uint {
	// Get the user ID from the middleware preset to avoid duplicate parsing
	if userID, exists := c.Get("bxUserID"); exists {
		return userID.(uint)
	}

	// Backup solution: parsing from claims (backward compatibility)
	if claims, exists := c.Get("bxClaims"); !exists {
		if cl, err := GetBxClaims(c); err != nil {
			return 0
		} else {
			return cl.BaseClaims.ID
		}
	} else {
		waitUse := claims.(*bxapiReq.CustomClaims)
		return waitUse.BaseClaims.ID
	}
}

// GetBxUserIDFast Fast get user ID (recommended to use)
// Get the user ID from the middleware preset to avoid duplicate parsing
func GetBxUserIDFast(c *gin.Context) uint {
	if userID, exists := c.Get("bxUserID"); exists {
		return userID.(uint)
	}
	return 0 // If it does not exist, return 0
}
func DxLoginToken(user dxapi.Login) (token string, claims bxapiReq.CustomClaims, err error) {
	j := NewDxJWT() // 使用专门的 BxJWT
	claims = j.CreateClaims(dxapiReq.BaseClaims{
		GetAgentId:         user.GetAgentId(),
		GetAgentname:       user.GetAgentname(),
		GetResponsibleName: user.GetResponsibleName(),
	})
	token, err = j.CreateToken(claims)
	if err != nil {
		return
	}
	return
}
