package support

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mo7amed4522/tayara_api/constants"
	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/model/common/response"
	supportReq "github.com/mo7amed4522/tayara_api/model/support/request"
	supportKafka "github.com/mo7amed4522/tayara_api/service/support/kafka"
	"github.com/mo7amed4522/tayara_api/service/support/stream"
	"github.com/mo7amed4522/tayara_api/utils"
	"go.uber.org/zap"
)

// MessageApi exposes message append, history and the SSE stream.
type MessageApi struct{}

// Send
//
//	@Tags		Support
//	@Summary	Append a message (body optional only when at least one attachment is claimed)
//	@Produce	application/json
//	@Param		id		path		int							true	"conversation id"
//	@Param		data	body		supportReq.SendMessageReq	true	"message payload"
//	@Success	200		{object}	response.Response
//	@Router		/api/v1/support/conversations/{id}/messages [post]
func (MessageApi) Send(c *gin.Context) {
	convID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || convID == 0 {
		response.FailWithError(response.ValidationErr("invalid conversation id"), c)
		return
	}
	var req supportReq.SendMessageReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithError(response.ValidationErr(err.Error()), c)
		return
	}
	view, err := supportService.MessageServices.Send(currentActor(c), uint(convID), req)
	if err != nil {
		respondErr(err, c)
		return
	}
	response.OkWithData(view, c)
}

// History
//
//	@Tags		Support
//	@Summary	Page a conversation's messages (supports sinceSeq replay)
//	@Produce	application/json
//	@Param		id		path		int							true	"conversation id"
//	@Param		data	query		supportReq.MessagePageReq	true	"page"
//	@Success	200		{object}	response.Response
//	@Router		/api/v1/support/conversations/{id}/messages [get]
func (MessageApi) History(c *gin.Context) {
	convID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || convID == 0 {
		response.FailWithError(response.ValidationErr("invalid conversation id"), c)
		return
	}
	var req supportReq.MessagePageReq
	_ = c.ShouldBindQuery(&req)
	list, total, totalPage, err := supportService.MessageServices.History(currentActor(c), uint(convID), req)
	if err != nil {
		respondErr(err, c)
		return
	}
	response.OkWithDetailed(response.PageResult{List: list, Total: total, TotalPage: totalPage, Page: req.Page, PageSize: req.PageSize}, "ok", c)
}

// Stream
//
//	@Tags		Support
//	@Summary	Server-sent events for a conversation (message.created + presence.changed)
//	@Produce	text/event-stream
//	@Param		id			path	int		true	"conversation id"
//	@Param		sinceSeq	query	int		false	"replay messages after this seq"
//	@Success	200			{string}	string	"event stream"
//	@Router		/api/v1/support/conversations/{id}/stream [get]
func (MessageApi) Stream(c *gin.Context) {
	convID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || convID == 0 {
		response.FailWithError(response.ValidationErr("invalid conversation id"), c)
		return
	}
	actor := currentActor(c)
	access, ok, err := supportService.AuthorizationServices.CanRead(actor, uint(convID))
	if err != nil {
		respondErr(err, c)
		return
	}
	if !ok {
		response.FailWithError(response.ForbiddenErr("not allowed to read this conversation"), c)
		return
	}

	heartbeat := global.GVA_CONFIG.Support.Presence.HeartbeatIntervalSeconds
	if heartbeat <= 0 {
		heartbeat = 15
	}
	connID := utils.NewUUIDv7()
	presence := supportService.PresenceServices
	presence.Connect(c.Request.Context(), actor, access.Conversation.ID, access.Conversation.TicketID, connID)
	defer presence.Disconnect(c.Request.Context(), actor.Type, actor.ID, connID, access.Conversation.ID, access.Conversation.TicketID)

	sub, cancel := stream.Default.Subscribe(access.Conversation.ID)
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

	// Replay from the database: SSE is a fast path, never the source of truth.
	sinceSeq, _ := strconv.ParseInt(c.Query("sinceSeq"), 10, 64)
	if replay, _, _, err := supportService.MessageServices.History(actor, access.Conversation.ID, supportReq.MessagePageReq{
		Page: 1, PageSize: 100, SinceSeq: sinceSeq,
	}); err == nil {
		for _, m := range replay {
			if payload, e := supportKafka.NewEnvelope(supportKafka.EventMessageCreated, access.Conversation.ID, access.Conversation.TicketID, m); e == nil {
				writeSSE(c, supportKafka.EventMessageCreated, payload)
			}
		}
	}

	ticker := time.NewTicker(time.Duration(heartbeat) * time.Second)
	defer ticker.Stop()
	authCheck := time.NewTicker(time.Minute)
	defer authCheck.Stop()
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
		case <-authCheck.C:
			if _, still, err := supportService.AuthorizationServices.CanRead(actor, access.Conversation.ID); err != nil || !still {
				writeSSE(c, "closing", []byte(`{"reason":"authorization revoked"}`))
				return
			}
		case <-ticker.C:
			presence.Heartbeat(ctx, actor.Type, actor.ID, connID)
			_, _ = fmt.Fprint(c.Writer, ": ping\n\n")
			flusher.Flush()
		}
	}
}

func writeSSE(c *gin.Context, eventType string, payload []byte) {
	w := c.Writer
	_, _ = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", eventType, payload)
	if f, ok := w.(interface{ Flush() }); ok {
		f.Flush()
	}
}

var (
	_ = constants.SupportPresenceOnline
	_ = zap.String
)
