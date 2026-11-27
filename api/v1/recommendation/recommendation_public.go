package recommendation

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mo7amed4522/tayara_api/model/common/response"
	"github.com/mo7amed4522/tayara_api/model/recommendation/request"
	"github.com/mo7amed4522/tayara_api/utils"
)

// RecommendationPublicApi customer-facing endpoints. They only ever read approved rows; the LLM
// is never called on a customer request path.
type RecommendationPublicApi struct{}

// List
//
//	@Tags		Recommendation
//	@Summary	Approved recommendations for a placement
//	@Produce	application/json
//	@Param		placement	query		string	true	"placement e.g. home_top"
//	@Param		audience	query		string	false	"audience segment"	default(all)
//	@Param		page		query		int		false	"page number"		default(1)
//	@Param		pageSize	query		int		false	"page size"			default(10)
//	@Success	200			{object}	response.Response
//	@Router		/api/v1/recommendation/public/list [get]
func (RecommendationPublicApi) List(c *gin.Context) {
	placement := c.Query("placement")
	audience := c.DefaultQuery("audience", "all")
	page := 1
	if raw := c.Query("page"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			page = n
		}
	}
	pageSize := 10
	if raw := c.Query("pageSize"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 && n <= 50 {
			pageSize = n
		}
	}
	list, err := recommendationSvc.GetPublicRecommendationList(placement, audience, page, pageSize)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(list, c)
}

// Event
//
//	@Tags		Recommendation
//	@Summary	Record an impression, click or booking for a recommendation
//	@Produce	application/json
//	@Param		data	body		request.RecommendationEventReq	true	"event"
//	@Success	200		{object}	response.Response
//	@Router		/api/v1/recommendation/public/event [post]
func (RecommendationPublicApi) Event(c *gin.Context) {
	var req request.RecommendationEventReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := recommendationSvc.RecordEvent(req, utils.GetUserID(c)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(c)
}
