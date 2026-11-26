package initialize

import (
	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/model/base"
	"github.com/mo7amed4522/tayara_api/model/booking"
	"github.com/mo7amed4522/tayara_api/model/bxapi"
	"github.com/mo7amed4522/tayara_api/model/catalog"
	"github.com/mo7amed4522/tayara_api/model/dxapi"
	"github.com/mo7amed4522/tayara_api/model/engage"
	paymentModel "github.com/mo7amed4522/tayara_api/model/payment"
	supportModel "github.com/mo7amed4522/tayara_api/model/support"
	"github.com/mo7amed4522/tayara_api/model/system"
)

func bizModel() error {
	db := global.GVA_DB
	err := db.AutoMigrate(
		// base / dictionaries
		&base.Country{},
		&base.City{},
		&base.Nationality{},
		&base.DocumentType{},
		&base.SystemSetting{},
		&base.PackageCategory{},
		&base.SettingChangeHistory{},
		// user
		&bxapi.BxDict{},
		&bxapi.BxsUserProfile{},
		// catalog
		&catalog.Package{},
		&catalog.PackageAvailability{},
		&catalog.PackageHotel{},
		&catalog.PackageActivity{},
		&catalog.PackageAddon{},
		&catalog.PackageRequiredField{},
		&catalog.PackageDocumentRequirement{},
		&catalog.PackageView{},
		&catalog.ItineraryDay{},
		&catalog.ItineraryItem{},
		&catalog.PackageInclusion{},
		&catalog.PackageExclusion{},
		&catalog.PackageHighlight{},
		&catalog.CustomTripTemplate{},
		&catalog.CustomTripTemplateCountry{},
		&catalog.CustomTripRequirementProfile{},
		// agent
		&dxapi.DxAgentProfile{},
		&dxapi.AgentPlan{},
		&dxapi.AgentDocument{},
		&dxapi.AgentNote{},
		// booking
		&booking.Booking{},
		&booking.BookingTraveler{},
		&booking.TravelerDocument{},
		&booking.BookingOffer{},
		&booking.BookingStatusHistory{},
		&booking.BookingDocument{},
		&booking.SLAEvent{},
		&booking.CustomTripRequest{},
		&booking.DocumentAccessLog{},
		&booking.Dispute{},
		// engage
		&engage.Notification{},
		&engage.NotificationTemplate{},
		&engage.Review{},
		&engage.AdCreative{},
		&engage.AdClick{},
		&engage.Wishlist{},
		&engage.Share{},
		&engage.Coupon{},
		&engage.DeviceToken{},
		&engage.PushToken{},
		&engage.PackageTopic{},
		&engage.ExploreTopic{},

		// payment
		&paymentModel.PaymentTransaction{},
		&paymentModel.Refund{},
		&paymentModel.Payout{},
		&paymentModel.CommissionRecord{},
		// support (tickets, chat, presence, attachments)
		&supportModel.Ticket{},
		&supportModel.Conversation{},
		&supportModel.Message{},
		&supportModel.AssignmentHistory{},
		&supportModel.Review{},
		&supportModel.Faq{},
		&supportModel.FaqVote{},
		&supportModel.Presence{},
		&supportModel.Attachment{},
		// system
		&system.SysAdminProfile{},
		&system.JwtBlacklist{},
		&system.PasswordReset{},
	)
	if err != nil {
		return err
	}
	return nil
}
