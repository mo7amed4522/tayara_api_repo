package middleware

import (
	"time"

	"github.com/mo7amed4522/tayara_api/constants"
	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/model/system"
	"github.com/mo7amed4522/tayara_api/utils"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type DomainTokenResult struct {
	Token     string
	ExpiresAt int64
}

func IssueDomainToken(user system.SysAdminProfile, headerName string) (DomainTokenResult, error) {
	token, claims, err := utils.LoginTokenWithAuthority(&user, constants.AuthoritySystem)
	if err != nil {
		global.GVA_LOG.Error("get token failed", zap.Error(err))
		return DomainTokenResult{}, err
	}
	if !global.GVA_CONFIG.System.UseMultipoint {
		return DomainTokenResult{Token: token, ExpiresAt: claims.RegisteredClaims.ExpiresAt.Unix() * 1000}, nil
	}

	redisKey := headerName + ":" + user.Username
	if jwtStr, err := jwtService.GetRedisJWT(redisKey); err == redis.Nil {
		if err := jwtService.SetRedisJWT(token, redisKey); err != nil {
			global.GVA_LOG.Error("set login status failed", zap.Error(err))
			return DomainTokenResult{}, err
		}
	} else if err != nil {
		global.GVA_LOG.Error("set login status failed", zap.Error(err))
		return DomainTokenResult{}, err
	} else {
		if err := jwtService.JsonInBlacklist(system.JwtBlacklist{Jwt: jwtStr}); err != nil {
			return DomainTokenResult{}, err
		}
		if err := jwtService.SetRedisJWT(token, redisKey); err != nil {
			return DomainTokenResult{}, err
		}
	}
	return DomainTokenResult{Token: token, ExpiresAt: claims.RegisteredClaims.ExpiresAt.Unix() * 1000}, nil
}

func SetDomainTokenCookie(host, headerName, token string, maxAge int, setCookie func(string, string, int, string, string, bool, bool)) {
	domain := utils.GetCookieDomain(host)
	setCookie(headerName, token, maxAge, "/", domain, false, false)
}

func DomainTokenMaxAge(expiresAtMs int64) int {
	return int(expiresAtMs/1000 - time.Now().Unix())
}
