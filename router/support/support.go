package support

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mo7amed4522/tayara_api/constants"
	"github.com/mo7amed4522/tayara_api/middleware"
)

// SupportRouter registers support endpoints across the customer (X-token) and staff (S-token)
// guard groups.
type SupportRouter struct{}

// InitSupportRouter mounts /api/v1/support for customers and staff.
func (r SupportRouter) InitSupportRouter(publicGroup, bxGroup, systemGroup *gin.RouterGroup) {
	public := publicGroup.Group("api/v1/support")
	customer := bxGroup.Group("api/v1/support")
	staff := systemGroup.Group("api/v1/support")
	// Endpoints both customers and staff use live on the public engine, so they are registered
	// once behind a guard that accepts either token; gin panics on a duplicate method+path.
	shared := publicGroup.Group("api/v1/support")
	shared.Use(middleware.SupportJWTAuth())

	// Unauthenticated page: a self-contained browser SSE client for manual/e2e verification.
	public.GET("/sse-client", sseClientApi.Client)

	// customer (X-token)
	{
		customer.POST("/tickets", ticketApi.Open)
		customer.GET("/tickets", ticketApi.ListOwn)
		customer.GET("/tickets/:id", ticketApi.Detail)
		customer.GET("/tickets/:id/faq", faqApi.List)
		customer.POST("/faqs/:id/vote", middleware.SupportRateLimit("faq-vote", 20, time.Minute), faqApi.Vote)
	}

	// customer + staff (X-token or S-token)
	{
		shared.GET("/conversations/:id", ticketApi.ConversationDetail)
		shared.GET("/conversations/:id/messages", messageApi.History)
		shared.POST("/conversations/:id/messages", middleware.SupportRateLimit("chat-send", 30, 10*time.Second), messageApi.Send)
		shared.GET("/conversations/:id/stream", messageApi.Stream)
		shared.GET("/presence", middleware.SupportRateLimit("presence-read", 60, time.Minute), presenceApi.Batch)
		shared.POST("/attachments", middleware.SupportRateLimit("attachment-upload", 20, time.Minute), attachmentApi.Upload)
		shared.GET("/attachments/:uuid", middleware.SupportRateLimit("attachment-read", 120, time.Minute), attachmentApi.Download)
	}

	// push token registration (X-token or S-token): both sides own devices
	r.InitSupportPushRouter(shared)

	// staff (S-token): support employees, quality reviewers and admins
	{
		staff.GET("/admin/tickets", ticketApi.AdminList)
		staff.GET("/admin/tickets/:id", ticketApi.AdminDetail)
		staff.POST("/admin/tickets/:id/assign", middleware.SupportRateLimit("assign", 60, time.Minute), adminApi.Assign)
		staff.GET("/admin/tickets/:id/assignments", adminApi.Assignments)
		staff.POST("/admin/tickets/:id/review", middleware.SupportRateLimit("review", 30, time.Minute), adminApi.Review)
		staff.GET("/admin/tickets/:id/reviews", adminApi.Reviews)
		staff.GET("/admin/chat-settings", adminApi.ChatSettings)
		staff.PUT("/admin/chat-settings", middleware.SupportRateLimit("settings", 30, time.Minute), adminApi.UpdateChatSettings)
		staff.GET("/admin/faqs", adminApi.ListFaqs)
		staff.POST("/admin/faqs", middleware.SupportRateLimit("faq-admin", 60, time.Minute), adminApi.UpsertFaq)
		staff.PUT("/admin/faqs", middleware.SupportRateLimit("faq-admin", 60, time.Minute), adminApi.UpsertFaq)
		staff.PUT("/admin/faqs/status", middleware.SupportRateLimit("faq-admin", 60, time.Minute), adminApi.FaqStatus)
	}
}

// InitSupportPushRouter mounts the device token endpoints on the shared (X-token or S-token)
// group. It must be called exactly once: gin panics on a duplicate method+path across groups.
func (r SupportRouter) InitSupportPushRouter(shared *gin.RouterGroup) {
	shared.POST("/push-tokens",
		middleware.SupportRateLimit(constants.SupportRateLimitPushTokenRegister, 20, time.Minute), pushApi.Register)
	shared.DELETE("/push-tokens/:platform",
		middleware.SupportRateLimit(constants.SupportRateLimitPushTokenRevoke, 20, time.Minute), pushApi.Revoke)
	// Real-time notification stream (the in-app half of the push pipeline).
	shared.GET("/notifications/stream",
		middleware.SupportRateLimit(constants.SupportRateLimitNotificationStream, 30, time.Minute), pushApi.NotificationStream)
}
