package initialize

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/plugin/email"
	"github.com/mo7amed4522/tayara_api/utils/plugin"
)

func PluginInit(group *gin.RouterGroup, Plugin ...plugin.Plugin) {
	for i := range Plugin {
		fmt.Println(Plugin[i].RouterPath(), "Registration started!")
		PluginGroup := group.Group(Plugin[i].RouterPath())
		Plugin[i].Register(PluginGroup)
		fmt.Println(Plugin[i].RouterPath(), "Registration successful!")
	}
}

func bizPluginV1(group ...*gin.RouterGroup) {
	private := group[0]
	public := group[1]
	// Adding a plugin that links permissions to a role: Example. Note the "import" option at the top in both local and online repository modes; you can switch between them as needed. The effect is the same.
	PluginInit(private, email.CreateEmailPlug(
		global.GVA_CONFIG.Email.To,
		global.GVA_CONFIG.Email.From,
		global.GVA_CONFIG.Email.Host,
		global.GVA_CONFIG.Email.Secret,
		global.GVA_CONFIG.Email.Nickname,
		global.GVA_CONFIG.Email.Port,
		global.GVA_CONFIG.Email.IsSSL,
	))
	holder(public, private)
}
