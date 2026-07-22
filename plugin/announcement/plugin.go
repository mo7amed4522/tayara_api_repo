package announcement

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/mo7amed4522/tayara_api/plugin/announcement/initialize"
	interfaces "github.com/mo7amed4522/tayara_api/utils/plugin/v2"
)

var _ interfaces.Plugin = (*plugin)(nil)

var Plugin = new(plugin)

type plugin struct{}

func (p *plugin) Register(group *gin.Engine) {
	ctx := context.Background()
	// If you need a configuration file, please fill in the configuration structure in config.Config, and fill in the key in the configuration file under the configuration file
	// initialize.Viper()
	// The automatically registered api data when installing the plugin should be implemented in the .Api method
	// The automatically registered api data when installing the plugin should be implemented in the .Menu method
	initialize.Gorm(ctx)
	initialize.Router(group)
}
