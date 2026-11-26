package support

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/mo7amed4522/tayara_api/model/common/response"
)

// respondErr maps an AppError to its truthful HTTP status and falls back to the legacy envelope
// for plain errors.
func respondErr(err error, c *gin.Context) {
	var appErr *response.AppError
	if errors.As(err, &appErr) {
		response.FailWithError(appErr, c)
		return
	}
	response.FailWithMessage(err.Error(), c)
}
