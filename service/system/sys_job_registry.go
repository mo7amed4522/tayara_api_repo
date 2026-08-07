package system

// ScheduledJob is a read-only view of a recurring background job the platform runs.
// It mirrors the timers wired in initialize/timer.go so an admin can see what runs and when.
type ScheduledJob struct {
	Name       string `json:"name"`       // job name
	Schedule   string `json:"schedule"`   // cadence description
	Enabled    bool   `json:"enabled"`    // always true while the server runs the ticker
	AutoAction string `json:"autoAction"` // what a breach/trigger does
}

// JobRegistry returns the platform's scheduled jobs (cron-registry view). This is the single
// place that documents the recurring timers; the actual scheduling lives in initialize/timer.go.
func (sysConfigService SystemConfigService) JobRegistry() []ScheduledJob {
	return []ScheduledJob{
		{
			Name:       "Booking SLA watchdog",
			Schedule:   "every 1 minute",
			Enabled:    true,
			AutoAction: "auto-cancel / offer-expiry / reminder on breached agent-response or user-payment SLA",
		},
		{
			Name:       "Reservation hold release",
			Schedule:   "every 1 minute",
			Enabled:    true,
			AutoAction: "move awaiting-payment bookings past their hold window to price_expired",
		},
		{
			Name:       "Ad creative scheduler",
			Schedule:   "every 1 minute",
			Enabled:    true,
			AutoAction: "flip scheduled -> active at start and active -> expired at end",
		},
	}
}
