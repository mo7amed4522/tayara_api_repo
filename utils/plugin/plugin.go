package plugin

import (
	"github.com/gin-gonic/gin"
)

const (
	OnlyFuncName = "Plugin"
)

// Plugin Plug-in mode interface
type Plugin interface {
	// Register Register Router
	Register(group *gin.RouterGroup)

	// RouterPath User returns to the registration route
	RouterPath() string
}
