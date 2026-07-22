package initialize

import (
	"github.com/gin-gonic/gin"
	"github.com/mo7amed4522/tayara_api/middleware"
)

func holder(routers ...*gin.RouterGroup) {
	_ = routers
}
func initBizRouter(routers ...*gin.RouterGroup) {
	privateGroup := routers[0]
	publicGroup := routers[1]
	bxPrivateGroup := routers[2]
	holder(publicGroup, privateGroup, bxPrivateGroup)

	bxPrivate := publicGroup.Group("")
	bxPrivate.Use(middleware.BxDomainJWTAuth())
	managePrivate := publicGroup.Group("")
	managePrivate.Use(middleware.ManageJWTAuth())
	systemPrivate := publicGroup.Group("")
	systemPrivate.Use(middleware.SystemJWTAuth())

}
