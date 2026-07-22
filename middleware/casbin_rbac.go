package middleware

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/model/common/response"
	"github.com/mo7amed4522/tayara_api/service"
	"github.com/mo7amed4522/tayara_api/utils"
)

var casbinService = service.ServiceGroupApp.SystemServiceGroup.CasbinService

// CasbinHandler Interceptor
func CasbinHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		waitUse, _ := utils.GetClaims(c)
		// Get the request PATH
		path := c.Request.URL.Path
		obj := strings.TrimPrefix(path, global.GVA_CONFIG.System.RouterPrefix)
		// Get the request method
		act := c.Request.Method
		// Get the user's role
		sub := strconv.Itoa(int(waitUse.AuthorityId))
		e := casbinService.Casbin() // Check if the strategy exists
		success, _ := e.Enforce(sub, obj, act)
		if !success {
			response.FailWithDetailed(gin.H{}, "Insufficient permissions", c)
			c.Abort()
			return
		}
		c.Next()
	}
}
