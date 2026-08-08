package search

import "github.com/gin-gonic/gin"

type SearchRouter struct{}

// InitSearchRouter registers the admin global-search endpoint on the system (admin) guard group.
func (r SearchRouter) InitSearchRouter(systemGroup *gin.RouterGroup) {
	system := systemGroup.Group("api/v1/admin")
	{
		system.GET("search", searchApi.GlobalSearch)
	}
}