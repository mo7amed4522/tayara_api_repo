package search

import (
	"github.com/gin-gonic/gin"
	"github.com/mo7amed4522/tayara_api/model/common/response"
)

type SearchApi struct{}

// GlobalSearch
//	@Tags		Search
//	@Summary	Admin global cross-entity search (users, agents, packages, bookings)
//	@Produce	application/json
//	@Param		keyword	query		string	true	"search keyword"
//	@Success	200		{object}	response.Response
//	@Router		/api/v1/admin/search [get]
func (SearchApi) GlobalSearch(c *gin.Context) {
	keyword := c.Query("keyword")
	res, err := searchService.SearchServices.Global(keyword)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(res, c)
}