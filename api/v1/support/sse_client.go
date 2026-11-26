package support

import (
	_ "embed"
	"net/http"

	"github.com/gin-gonic/gin"
)

// sseClientHTML is a self-contained browser client for exercising the SSE stream end to end:
// it sets the x-token cookie, opens EventSource, renders message.created / presence.changed
// frames and can POST a message into the same conversation.
//
//go:embed sse_client.html
var sseClientHTML []byte

// SseClientApi serves the browser SSE session page.
type SseClientApi struct{}

// Client
//
//	@Tags		Support
//	@Summary	Browser page that opens a real SSE session for a conversation
//	@Produce	html
//	@Success	200	{string}	string	"SSE client page"
//	@Router		/api/v1/support/sse-client [get]
func (SseClientApi) Client(c *gin.Context) {
	c.Data(http.StatusOK, "text/html; charset=utf-8", sseClientHTML)
}
