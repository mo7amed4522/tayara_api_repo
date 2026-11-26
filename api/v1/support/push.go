package support

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mo7amed4522/tayara_api/model/common/response"
	supportReq "github.com/mo7amed4522/tayara_api/model/support/request"
	"github.com/mo7amed4522/tayara_api/service/support/stream"
)

// PushApi exposes device token registration for the offline delivery path and the real-time
// notification stream.
type PushApi struct{}

// Register
//
//	@Tags		Support
//	@Summary	Register (or refresh) a device push token for the caller
//	@Produce	application/json
//	@Param		data	body		supportReq.RegisterPushTokenReq	true	"platform, token, appVersion"
//	@Success	200		{object}	response.Response
//	@Router		/api/v1/support/push-tokens [post]
func (PushApi) Register(c *gin.Context) {
	actor := currentActor(c)
	if !actor.Valid() {
		response.FailWithError(response.ForbiddenErr("not authenticated"), c)
		return
	}
	var req supportReq.RegisterPushTokenReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithError(response.ValidationErr(err.Error()), c)
		return
	}
	row, err := supportService.PushServices.RegisterToken(actor.Type, actor.ID, req)
	if err != nil {
		respondErr(err, c)
		return
	}
	response.OkWithData(gin.H{"platform": row.Platform, "createdAt": row.CreatedAt}, c)
}

// Revoke
//
//	@Tags		Support
//	@Summary	Revoke the caller's active push tokens on one platform
//	@Produce	application/json
//	@Param		platform	path		string	true	"ios|android|web"
//	@Success	200			{object}	response.Response
//	@Router		/api/v1/support/push-tokens/{platform} [delete]
func (PushApi) Revoke(c *gin.Context) {
	actor := currentActor(c)
	if !actor.Valid() {
		response.FailWithError(response.ForbiddenErr("not authenticated"), c)
		return
	}
	req := supportReq.RevokePushTokenReq{Platform: c.Param("platform")}
	if err := supportService.PushServices.RevokeToken(actor.Type, actor.ID, req); err != nil {
		respondErr(err, c)
		return
	}
	response.OkWithMessage("push tokens revoked", c)
}

// NotificationStream
//
//	@Tags		Support
//	@Summary	Server-sent events for the caller's own support notifications (notification.created)
//	@Produce	text/event-stream
//	@Success	200	{string}	string	"event stream"
//	@Router		/api/v1/support/notifications/stream [get]
//
// It is the in-app half of the push pipeline: the same recorded notification that goes out as an
// FCM push is also written here, so an open app updates its notification list without polling.
func (PushApi) NotificationStream(c *gin.Context) {
	actor := currentActor(c)
	if !actor.Valid() {
		response.FailWithError(response.ForbiddenErr("not authenticated"), c)
		return
	}
	sub, cancel := stream.Default.SubscribeActor(actor.Type + ":" + strconv.FormatUint(uint64(actor.ID), 10))
	defer cancel()

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
	c.Status(200)
	flusher, ok := c.Writer.(gin.ResponseWriter)
	if !ok {
		return
	}
	flusher.Flush()
	writeSSE(c, "connected", []byte(`{"ok":true}`))

	ticker := time.NewTicker(25 * time.Second)
	defer ticker.Stop()
	ctx := c.Request.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case ev, open := <-sub.C:
			if !open {
				return
			}
			writeSSE(c, ev.Type, ev.Data)
		case <-ticker.C:
			_, _ = fmt.Fprint(c.Writer, ": ping\n\n")
			flusher.Flush()
		}
	}
}
