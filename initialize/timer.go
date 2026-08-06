package initialize

import (
	"time"

	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/service/booking"
	"github.com/mo7amed4522/tayara_api/service/engage"
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
