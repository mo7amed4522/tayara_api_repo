package analytics

import (
	"errors"
	"time"

	"github.com/mo7amed4522/tayara_api/constants"
	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/model/analytics/response"
	bookingModel "github.com/mo7amed4522/tayara_api/model/booking"
	"github.com/mo7amed4522/tayara_api/model/bxapi"
	"github.com/mo7amed4522/tayara_api/model/catalog"
	"github.com/mo7amed4522/tayara_api/model/dxapi"
	"github.com/mo7amed4522/tayara_api/model/engage"
	paymentModel "github.com/mo7amed4522/tayara_api/model/payment"
)

type AnalyticsServices struct{}

// AdminUsersCard aggregates the admin Users dashboard card. Admin-scoped read.
func (a AnalyticsServices) AdminUsersCard() (*response.AdminUsersCard, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	out := &response.AdminUsersCard{}
	_ = global.GVA_DB.Model(&bxapi.BxsUserProfile{}).Count(&out.TotalUsers).Error
	_ = global.GVA_DB.Model(&bxapi.BxsUserProfile{}).
		Where("created_at >= ?", time.Now().Add(-7*24*time.Hour)).Count(&out.NewUsersThisWeek).Error
	_ = global.GVA_DB.Model(&bxapi.BxsUserProfile{}).
		Select("COALESCE(SUM(total_spend),0)").Scan(&out.TotalSpend).Error
	_ = global.GVA_DB.Model(&bxapi.BxsUserProfile{}).
		Where("kyc_status = ?", constants.KYCStatusApproved).Count(&out.KYCApproved).Error
	_ = global.GVA_DB.Model(&bookingModel.Booking{}).
		Where("booking_status NOT IN ?", []string{
			constants.BookingStatusCompleted, constants.BookingStatusCancelled, constants.BookingStatusRefunded,
		}).Count(&out.ActiveBookings).Error
	_ = global.GVA_DB.Model(&engage.Share{}).Count(&out.Shares).Error
	_ = global.GVA_DB.Model(&engage.Review{}).Count(&out.Reviews).Error
	return out, nil
}

// AdminAgentsCard aggregates the admin dashboard agents card.
func (a AnalyticsServices) AdminAgentsCard() (*response.AdminAgentsCard, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	out := &response.AdminAgentsCard{}
	_ = global.GVA_DB.Model(&dxapi.DxAgentProfile{}).Count(&out.TotalAgents).Error
	_ = global.GVA_DB.Model(&dxapi.DxAgentProfile{}).Where("status = ?", "pending").Count(&out.Pending).Error
	_ = global.GVA_DB.Model(&dxapi.DxAgentProfile{}).Where("status = ?", "active").Count(&out.Active).Error
	_ = global.GVA_DB.Model(&dxapi.DxAgentProfile{}).Where("status = ?", "suspended").Count(&out.Suspended).Error
	_ = global.GVA_DB.Model(&dxapi.DxAgentProfile{}).Where("status = ?", "rejected").Count(&out.Rejected).Error
	_ = global.GVA_DB.Model(&catalog.Package{}).Where("status = ?", constants.PackageStatusActive).
		Count(&out.TotalPackages).Error
	return out, nil
}

// AdminBusinessCard aggregates the admin dashboard business card (money = minor units).
func (a AnalyticsServices) AdminBusinessCard() (*response.AdminBusinessCard, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	out := &response.AdminBusinessCard{}
	_ = global.GVA_DB.Model(&bookingModel.Booking{}).
		Select("COALESCE(SUM(amount_gross - COALESCE(discount_amount,0)),0)").Scan(&out.GMV).Error
	_ = global.GVA_DB.Model(&paymentModel.PaymentTransaction{}).
		Where("status = ?", constants.PaymentLifecycleCaptured).
		Select("COALESCE(SUM(commission_amount),0)").Scan(&out.Commission).Error
	_ = global.GVA_DB.Model(&paymentModel.PaymentTransaction{}).
		Where("status = ?", constants.PaymentLifecycleCaptured).
		Select("COALESCE(SUM(fee_amount),0)").Scan(&out.Fees).Error
	_ = global.GVA_DB.Model(&paymentModel.Payout{}).
		Where("payout_status = ?", constants.PayoutStatusPaid).
		Select("COALESCE(SUM(agent_share),0)").Scan(&out.Payouts).Error
	_ = global.GVA_DB.Model(&paymentModel.Refund{}).
		Where("status = ?", "refunded").Select("COALESCE(SUM(amount),0)").Scan(&out.Refunds).Error
	out.NetProfit = out.Commission + out.Fees - out.Refunds
	return out, nil
}

// AgentCard aggregates a single agent's dashboard metrics. Agent row-scoping is enforced here:
// the calling agent id is baked into every query, so one agent can never read another's data.
func (a AnalyticsServices) AgentCard(agentID uint) (*response.AgentCard, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	var profile dxapi.DxAgentProfile
	if err := global.GVA_DB.First(&profile, agentID).Error; err != nil {
		return nil, errors.New("agent not found")
	}
	out := &response.AgentCard{AgentID: agentID, SLAHours: profile.SLAHours, SLAEnabled: profile.SLAHours > 0}

	_ = global.GVA_DB.Model(&bookingModel.Booking{}).Where("agent_id = ?", agentID).
		Count(&out.TotalBookings).Error
	_ = global.GVA_DB.Model(&bookingModel.Booking{}).
		Where("agent_id = ? AND booking_status NOT IN ?", agentID, []string{
			constants.BookingStatusCompleted, constants.BookingStatusCancelled, constants.BookingStatusRefunded,
		}).Count(&out.ActiveBookings).Error
	_ = global.GVA_DB.Model(&bookingModel.Booking{}).
		Where("agent_id = ? AND booking_status = ?", agentID, constants.BookingStatusCompleted).
		Count(&out.CompletedBookings).Error
	_ = global.GVA_DB.Model(&paymentModel.PaymentTransaction{}).
		Where("agent_id = ? AND status = ?", agentID, constants.PaymentLifecycleCaptured).
		Select("COALESCE(SUM(agent_share),0)").Scan(&out.Revenue).Error
	_ = global.GVA_DB.Model(&paymentModel.Payout{}).
		Where("agent_id = ? AND payout_status = ?", agentID, constants.PayoutStatusPending).
		Select("COALESCE(SUM(agent_share),0)").Scan(&out.PendingPayouts).Error
	_ = global.GVA_DB.Model(&catalog.Package{}).
		Where("agent_id = ?", agentID).Count(&out.TotalPackages).Error
	_ = global.GVA_DB.Model(&catalog.PackageView{}).
		Joins("JOIN catalog_package p ON p.id = catalog_package_view.package_id").
		Where("p.agent_id = ?", agentID).Count(&out.PackageViews).Error

	// avg confirmation time = mean of (price_confirmed_at - requested_at) across confirmed bookings
	var rows []struct {
		RequestedAt *time.Time
		ConfirmedAt *time.Time
	}
	_ = global.GVA_DB.Model(&bookingModel.Booking{}).
		Where("agent_id = ? AND price_confirmed_at IS NOT NULL AND requested_at IS NOT NULL", agentID).
		Select("requested_at, price_confirmed_at AS confirmed_at").Scan(&rows).Error
	if len(rows) > 0 {
		var total float64
		for _, r := range rows {
			if r.RequestedAt != nil && r.ConfirmedAt != nil {
				total += r.ConfirmedAt.Sub(*r.RequestedAt).Hours()
			}
		}
		out.AvgConfirmationHours = total / float64(len(rows))
	}
	return out, nil
}
