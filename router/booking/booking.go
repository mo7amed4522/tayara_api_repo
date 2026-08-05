package booking

import "github.com/gin-gonic/gin"

type BookingRouter struct{}

// InitBookingRouter registers booking endpoints across the three guard groups:
// public (buyer), manage (agent), system (admin).
func (r BookingRouter) InitBookingRouter(publicGroup, manageGroup, systemGroup *gin.RouterGroup) {
	user := publicGroup.Group("api/v1/booking")
	manage := manageGroup.Group("api/v1/booking/manage")
	system := systemGroup.Group("api/v1/booking/admin")

	{
		user.POST("", bookingApi.CreateBooking)
		user.GET("", bookingApi.ListOwnBookings)
		user.GET("detail", bookingApi.GetOwnBooking)
		user.POST("cancel", bookingApi.CancelBooking)
		user.POST("documents", docApi.SubmitDocuments)
		user.GET("documents", docApi.GetUserDocuments)
	}
	{
		manage.GET("", bookingApi.ListAgentBookings)
		manage.GET("detail", bookingApi.GetAgentBooking)
		manage.POST("review", bookingApi.ReviewBooking)
		manage.POST("offer", bookingApi.SubmitOffer)
		manage.POST("reject", bookingApi.RejectBooking)
		manage.GET("documents", docApi.AgentDocuments)
		manage.POST("documents/upload", docApi.UploadBookingDocument)
		manage.POST("documents/review", docApi.ApproveDocument)
		manage.POST("process", docApi.MarkProcessing)
		manage.POST("issue", docApi.IssueBooking)
		manage.POST("complete", docApi.CompleteBooking)
	}
	{
		system.GET("", bookingApi.ListAllBookings)
		system.POST("force-cancel", bookingApi.ForceCancel)
		system.POST("kyc", docApi.AdminReviewKYC)
	}
}
