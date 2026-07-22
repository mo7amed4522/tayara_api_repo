package system

import (
	"context"

	"go.uber.org/zap"

	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/model/system"
	"github.com/mo7amed4522/tayara_api/utils"
	"github.com/redis/go-redis/v9"
)

type JwtService struct{}

var JwtServiceApp = new(JwtService)

//@author: [mo7amed4522](https://github.com/mo7amed4522)
//@function: JsonInBlacklist
//@description: Blacklist jwt
//@param: jwtList model.JwtBlacklist
//@return: err error

func (jwtService *JwtService) JsonInBlacklist(jwtList system.JwtBlacklist) (err error) {
	err = global.GVA_DB.Create(&jwtList).Error
	if err != nil {
		return
	}
	global.BlackCache.SetDefault(jwtList.Jwt, struct{}{})
	return
}

//@author: [mo7amed4522](https://github.com/mo7amed4522)
//@function: IsBlacklist
//@description: Check if JWT is in the blacklist
//@param: jwt string
//@return: bool

func (jwtService *JwtService) IsBlacklist(jwt string) bool {
	_, ok := global.BlackCache.Get(jwt)
	return ok
	// err := global.GVA_DB.Where("jwt = ?", jwt).First(&system.JwtBlacklist{}).Error
	// isNotFound := errors.Is(err, gorm.ErrRecordNotFound)
	// return !isNotFound
}

//@author: [mo7amed4522](https://github.com/mo7amed4522)
//@function: GetRedisJWT
//@description: Get jwt from redis
//@param: userName string
//@return: redisJWT string, err error

func (jwtService *JwtService) GetRedisJWT(userName string) (redisJWT string, err error) {
	// Try the provided key first
	redisJWT, err = global.GVA_REDIS.Get(context.Background(), userName).Result()
	if err == nil {
		return redisJWT, nil
	}
	if err != redis.Nil {
		return "", err
	}

	// If not found, try common header-prefixed keys used by domain tokens
	prefixes := []string{"S-token:", "A-token:", "X-token:"}
	for _, p := range prefixes {
		v, e := global.GVA_REDIS.Get(context.Background(), p+userName).Result()
		if e == nil {
			return v, nil
		}
		if e != redis.Nil {
			return "", e
		}
	}
	return "", redis.Nil
}

//@author: [mo7amed4522](https://github.com/mo7amed4522)
//@function: SetRedisJWT
//@description: Set jwt in redis and set expiration time
//@param: jwt string, userName string
//@return: err error

func (jwtService *JwtService) SetRedisJWT(jwt string, userName string) (err error) {
	// The expiration time here is equal to the jwt expiration time
	dr, err := utils.ParseDuration(global.GVA_CONFIG.JWT.ExpiresTime)
	if err != nil {
		return err
	}
	timer := dr
	err = global.GVA_REDIS.Set(context.Background(), userName, jwt, timer).Err()
	return err
}

func LoadAll() {
	var data []string
	err := global.GVA_DB.Model(&system.JwtBlacklist{}).Select("jwt").Find(&data).Error
	if err != nil {
		global.GVA_LOG.Error("Failed to load database jwt blacklist!", zap.Error(err))
		return
	}
	for i := 0; i < len(data); i++ {
		global.BlackCache.SetDefault(data[i], struct{}{})
	} // jwt blacklist added to BlackCache
}
