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
		&catalog.CustomTripTemplate{},
		&catalog.CustomTripTemplateCountry{},
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
		// engage
		&engage.Notification{},
		// payment
		&paymentModel.PaymentTransaction{},
		&paymentModel.Refund{},
		&paymentModel.Payout{},
		// system
		&system.SysAdminProfile{},
		&system.JwtBlacklist{},
	)
	if err != nil {
		return err
	}
	return nil
}
