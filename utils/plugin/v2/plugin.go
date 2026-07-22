package plugin

import (
	"github.com/gin-gonic/gin"
)

// Plugin Plug-in mode interface v2
type Plugin interface {
	// Register Register Router
	Register(group *gin.Engine)
}
