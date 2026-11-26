package support

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/model/common/response"
)

// PresenceApi exposes the batch presence read.
type PresenceApi struct{}

// Batch
//
//	@Tags		Support
//	@Summary	Presence + last seen for actors sharing a readable conversation
//	@Produce	application/json
//	@Param		ids	query		string	true	"comma separated actor keys, e.g. support:9,customer:51"
//	@Success	200	{object}	response.Response
//	@Router		/api/v1/support/presence [get]
func (PresenceApi) Batch(c *gin.Context) {
	raw := c.Query("ids")
	if strings.TrimSpace(raw) == "" {
		response.FailWithError(response.ValidationErr("ids is required"), c)
		return
	}
	limit := global.GVA_CONFIG.Support.Presence.BatchReadMaxIds
	if limit <= 0 {
		limit = 50
	}
	parts := strings.Split(raw, ",")
	if len(parts) > limit {
		response.FailWithError(response.ValidationErr("too many ids"), c)
		return
	}
	views, err := supportService.PresenceServices.BatchRead(c.Request.Context(), currentActor(c), parts)
	if err != nil {
		respondErr(err, c)
		return
	}
	response.OkWithData(views, c)
}
