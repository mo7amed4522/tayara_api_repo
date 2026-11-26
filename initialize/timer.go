package initialize

import (
	"time"

	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/service/booking"
	"github.com/mo7amed4522/tayara_api/service/engage"
	supportService "github.com/mo7amed4522/tayara_api/service/support"
	supportPush "github.com/mo7amed4522/tayara_api/service/support/push"
	"go.uber.org/zap"
)

func Timer() {
	// Booking SLA watchdog: every minute scan for breached SLA timers (agent response /
	// user payment) and apply the configured auto action through the guarded transition.
	go func() {
		if global.GVA_DB == nil {
			return
		}
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		var svc booking.BookingServices
		for range ticker.C {
			if err := svc.ProcessOverdueSLA(); err != nil {
				global.GVA_LOG.Warn("booking SLA sweep failed", zap.Error(err))
			}
			// release reservation holds that outlived their pay window
			if err := svc.ReleaseExpiredHolds(); err != nil {
				global.GVA_LOG.Warn("hold release sweep failed", zap.Error(err))
			}
		}
	}()
	// Support presence expiry: repair state left behind by a killed process (TTL sweep).
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		var svc supportService.PresenceServices
		for range ticker.C {
			if err := svc.SweepExpired(); err != nil {
				global.GVA_LOG.Warn("support presence sweep failed", zap.Error(err))
			}
		}
	}()
	// Support last-seen flush + orphan attachment sweep.
	go func() {
		if global.GVA_DB == nil {
			return
		}
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		var lastSeen supportService.LastSeenServices
		var attachments supportService.AttachmentServices
		var outbox supportService.OutboxServices
		for range ticker.C {
			if err := lastSeen.Flush(); err != nil {
				global.GVA_LOG.Warn("support last seen flush failed", zap.Error(err))
			}
			if err := attachments.OrphanCleanup(); err != nil {
				global.GVA_LOG.Warn("support orphan attachment sweep failed", zap.Error(err))
			}
			if err := outbox.Sweep(); err != nil {
				global.GVA_LOG.Warn("support outbox sweep failed", zap.Error(err))
			}
		}
	}()
	// Support push retry: re-attempt failed dispatches until the attempt cap (append-only).
	go func() {
		if global.GVA_DB == nil || !global.GVA_CONFIG.Support.Push.Enabled {
			return
		}
		ticker := time.NewTicker(2 * time.Minute)
		defer ticker.Stop()
		var svc supportPush.PushServices
		for range ticker.C {
			if err := svc.RetryFailed(); err != nil {
				global.GVA_LOG.Warn("support push retry sweep failed", zap.Error(err))
			}
		}
	}()
	// Ad scheduler: flip scheduled -> active at start_datetime and active -> expired at end.
	go func() {
		if global.GVA_DB == nil {
			return
		}
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		var ads engage.AdServices
		for range ticker.C {
			if err := ads.RunScheduler(); err != nil {
				global.GVA_LOG.Warn("ad scheduler failed", zap.Error(err))
			}
		}
	}()
}
