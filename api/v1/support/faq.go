package support

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mo7amed4522/tayara_api/model/common/response"
	supportReq "github.com/mo7amed4522/tayara_api/model/support/request"
	"github.com/mo7amed4522/tayara_api/utils"
)

// FaqApi exposes the pre-chat Q&A and voting.
type FaqApi struct{}

// List
//
//	@Tags		Support
//	@Summary	Pre-chat FAQ list with tallies and the caller's vote
//	@Produce	application/json
//	@Param		id	path		int	true	"ticket id"
//	@Success	200	{object}	response.Response
//	@Router		/api/v1/support/tickets/{id}/faq [get]
func (FaqApi) List(c *gin.Context) {
	views, err := supportService.FaqServices.List(utils.GetBxUserID(c))
	if err != nil {
		respondErr(err, c)
		return
	}
	response.OkWithData(views, c)
}

// Vote
//
//	@Tags		Support
//	@Summary	Vote on a FAQ entry (change-of-mind appends a new record)
//	@Produce	application/json
//	@Param		id		path		int					true	"faq id"
//	@Param		data	body		supportReq.VoteReq	true	"vote"
//	@Success	200		{object}	response.Response
//	@Router		/api/v1/support/faqs/{id}/vote [post]
func (FaqApi) Vote(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithError(response.ValidationErr("invalid faq id"), c)
		return
	}
	var req supportReq.VoteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithError(response.ValidationErr(err.Error()), c)
		return
	}
	if err := supportService.FaqServices.Vote(utils.GetBxUserID(c), uint(id), req); err != nil {
		respondErr(err, c)
		return
	}
	response.OkWithMessage("vote recorded", c)
}
