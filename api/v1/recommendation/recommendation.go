package recommendation

import (
	"strconv"

	"github.com/gin-gonic/gin"
	commonReq "github.com/mo7amed4522/tayara_api/model/common/request"
	"github.com/mo7amed4522/tayara_api/model/common/response"
	"github.com/mo7amed4522/tayara_api/model/recommendation/request"
	"github.com/mo7amed4522/tayara_api/utils"
)

// RecommendationApi admin endpoints for stored recommendations, settings and audit runs.
type RecommendationApi struct{}

// CreateRecommendation
//
//	@Tags		Recommendation
//	@Summary	Create a manual recommendation (admin)
//	@Security	ApiKeyAuth
//	@Produce	application/json
//	@Param		data	body		request.CreateRecommendationReq	true	"recommendation"
//	@Success	200		{object}	response.Response
//	@Router		/api/v1/recommendation/createRecommendation [post]
func (RecommendationApi) CreateRecommendation(c *gin.Context) {
	var req request.CreateRecommendationReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	rec, err := recommendationSvc.CreateRecommendation(req, utils.GetUserID(c))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(rec, "Recommendation created", c)
}

// UpdateRecommendation
//
//	@Tags		Recommendation
//	@Summary	Update the editable fields of a recommendation (admin)
//	@Security	ApiKeyAuth
//	@Produce	application/json
//	@Param		data	body		request.UpdateRecommendationReq	true	"recommendation"
//	@Success	200		{object}	response.Response
//	@Router		/api/v1/recommendation/updateRecommendation [put]
func (RecommendationApi) UpdateRecommendation(c *gin.Context) {
	var req request.UpdateRecommendationReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := recommendationSvc.UpdateRecommendation(req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("Recommendation updated", c)
}

// DeleteRecommendation
//
//	@Tags		Recommendation
//	@Summary	Delete one recommendation (admin)
//	@Security	ApiKeyAuth
//	@Produce	application/json
//	@Param		id	query		int	true	"recommendation id"
//	@Success	200	{object}	response.Response
//	@Router		/api/v1/recommendation/deleteRecommendation [delete]
func (RecommendationApi) DeleteRecommendation(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("id is required", c)
		return
	}
	if err := recommendationSvc.DeleteRecommendation(uint(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("Recommendation deleted", c)
}

// DeleteRecommendationByIds
//
//	@Tags		Recommendation
//	@Summary	Delete several recommendations (admin)
//	@Security	ApiKeyAuth
//	@Produce	application/json
//	@Param		data	body		commonReq.IdsReq	true	"ids"
//	@Success	200		{object}	response.Response
//	@Router		/api/v1/recommendation/deleteRecommendationByIds [delete]
func (RecommendationApi) DeleteRecommendationByIds(c *gin.Context) {
	var req commonReq.IdsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	ids := make([]uint, 0, len(req.Ids))
	for _, id := range req.Ids {
		if id > 0 {
			ids = append(ids, uint(id))
		}
	}
	if err := recommendationSvc.DeleteRecommendationByIds(ids); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("Recommendations deleted", c)
}

// FindRecommendation
//
//	@Tags		Recommendation
//	@Summary	Get one recommendation (admin)
//	@Security	ApiKeyAuth
//	@Produce	application/json
//	@Param		id	query		int	true	"recommendation id"
//	@Success	200	{object}	response.Response
//	@Router		/api/v1/recommendation/findRecommendation [get]
func (RecommendationApi) FindRecommendation(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("id is required", c)
		return
	}
	rec, err := recommendationSvc.FindRecommendation(uint(id))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(rec, c)
}

// GetRecommendationList
//
//	@Tags		Recommendation
//	@Summary	Paginated recommendation list with filters (admin)
//	@Security	ApiKeyAuth
//	@Produce	application/json
//	@Param		page		query		int		false	"page"			default(1)
//	@Param		pageSize	query		int		false	"page size"		default(10)
//	@Param		targetType	query		string	false	"package | trip | agency"
//	@Param		placement	query		string	false	"placement"
//	@Param		audience	query		string	false	"audience"
//	@Param		status		query		string	false	"pending | approved | rejected | expired"
//	@Param		source		query		string	false	"llm | rules | admin"
//	@Param		keyword		query		string	false	"reason keyword"
//	@Success	200			{object}	response.Response
//	@Router		/api/v1/recommendation/getRecommendationList [get]
func (RecommendationApi) GetRecommendationList(c *gin.Context) {
	var info request.RecommendationSearch
	if err := c.ShouldBindQuery(&info); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := recommendationSvc.GetRecommendationList(info)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(utils.BuildPaginateResponse(list, total, info.Page, info.PageSize), "ok", c)
}

// Review
//
//	@Tags		Recommendation
//	@Summary	Approve or reject one recommendation (admin)
//	@Security	ApiKeyAuth
//	@Produce	application/json
//	@Param		data	body		request.ReviewRecommendationReq	true	"review"
//	@Success	200		{object}	response.Response
//	@Router		/api/v1/recommendation/review [post]
func (RecommendationApi) Review(c *gin.Context) {
	var req request.ReviewRecommendationReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := recommendationSvc.ReviewRecommendation(req.ID, utils.GetUserID(c), req.Action); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("Recommendation reviewed", c)
}

// BatchReview
//
//	@Tags		Recommendation
//	@Summary	Approve or reject several recommendations (admin)
//	@Security	ApiKeyAuth
//	@Produce	application/json
//	@Param		data	body		request.BatchReviewReq	true	"batch review"
//	@Success	200		{object}	response.Response
//	@Router		/api/v1/recommendation/batchReview [post]
func (RecommendationApi) BatchReview(c *gin.Context) {
	var req request.BatchReviewReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := recommendationSvc.BatchReview(req.IDs, utils.GetUserID(c), req.Action); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("Recommendations reviewed", c)
}

// Generate
//
//	@Tags		Recommendation
//	@Summary	Trigger an LLM recommendation run for a placement (admin)
//	@Security	ApiKeyAuth
//	@Produce	application/json
//	@Param		data	body		request.GenerateReq	true	"placement + audience"
//	@Success	200		{object}	response.Response
//	@Router		/api/v1/recommendation/generate [post]
func (RecommendationApi) Generate(c *gin.Context) {
	var req request.GenerateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	run, err := recommendationSvc.TriggerGenerate(req.Placement, req.Audience, utils.GetUserID(c))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(run, "Generation finished", c)
}

// GetRunList
//
//	@Tags		Recommendation
//	@Summary	Paginated audit list of generation runs (admin)
//	@Security	ApiKeyAuth
//	@Produce	application/json
//	@Param		page		query		int		false	"page"		default(1)
//	@Param		pageSize	query		int		false	"page size"	default(10)
//	@Param		placement	query		string	false	"placement"
//	@Param		status		query		string	false	"running | success | failed | fallback"
//	@Success	200			{object}	response.Response
//	@Router		/api/v1/recommendation/getRunList [get]
func (RecommendationApi) GetRunList(c *gin.Context) {
	var req request.RunSearch
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := recommendationSvc.GetRunList(req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(utils.BuildPaginateResponse(list, total, req.Page, req.PageSize), "ok", c)
}

// FindRun
//
//	@Tags		Recommendation
//	@Summary	Get one generation run with its prompt snapshot (admin)
//	@Security	ApiKeyAuth
//	@Produce	application/json
//	@Param		id	query		int	true	"run id"
//	@Success	200	{object}	response.Response
//	@Router		/api/v1/recommendation/findRun [get]
func (RecommendationApi) FindRun(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("id is required", c)
		return
	}
	run, err := recommendationSvc.FindRun(uint(id))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(run, c)
}

// GetStats
//
//	@Tags		Recommendation
//	@Summary	Impressions, clicks, CTR and bookings per recommendation (admin)
//	@Security	ApiKeyAuth
//	@Produce	application/json
//	@Param		placement	query		string	false	"placement"
//	@Param		startTime	query		string	false	"YYYY-MM-DD"
//	@Param		endTime		query		string	false	"YYYY-MM-DD"
//	@Success	200			{object}	response.Response
//	@Router		/api/v1/recommendation/getStats [get]
func (RecommendationApi) GetStats(c *gin.Context) {
	var req request.BaseTimeRange
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	from, to, err := recommendationSvc.ParseRange(req.StartTime, req.EndTime)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	stats, err := recommendationSvc.GetStats(req.Placement, from, to)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(stats, c)
}

// GetSetting
//
//	@Tags		Recommendation
//	@Summary	Read the engine setting for a placement (admin)
//	@Security	ApiKeyAuth
//	@Produce	application/json
//	@Param		placement	query		string	true	"placement"
//	@Success	200			{object}	response.Response
//	@Router		/api/v1/recommendation/getSetting [get]
func (RecommendationApi) GetSetting(c *gin.Context) {
	setting, err := recommendationSvc.GetSetting(c.Query("placement"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(setting, c)
}

// UpdateSetting
//
//	@Tags		Recommendation
//	@Summary	Update the engine setting for a placement (admin)
//	@Security	ApiKeyAuth
//	@Produce	application/json
//	@Param		data	body		request.UpdateSettingReq	true	"setting"
//	@Success	200		{object}	response.Response
//	@Router		/api/v1/recommendation/updateSetting [put]
func (RecommendationApi) UpdateSetting(c *gin.Context) {
	var req request.UpdateSettingReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	setting, err := recommendationSvc.UpdateSetting(req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(setting, "Setting updated", c)
}
