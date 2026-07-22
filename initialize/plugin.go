package initialize

import (
	"github.com/gin-gonic/gin"
	"github.com/mo7amed4522/tayara_api/global"
)

func InstallPlugin(PrivateGroup *gin.RouterGroup, PublicRouter *gin.RouterGroup, engine *gin.Engine) {
	if global.GVA_DB == nil {
		global.GVA_LOG.Info("project is not initialized, cannot install plugins, initialize the project again to complete the plugin installation")
		return
	}
	bizPluginV1(PrivateGroup, PublicRouter)
	bizPluginV2(engine)
}
