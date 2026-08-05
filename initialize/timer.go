package initialize

import (
	"time"

	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/service/booking"
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
		}
	}()
}
