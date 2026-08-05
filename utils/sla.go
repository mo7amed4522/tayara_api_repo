package utils

import "time"

// SLA clock helpers. SLA windows and hold durations are hours read from runtime System Settings;
// these helpers convert them to concrete deadlines and evaluate breaches. Kept as pure functions
// so SLA logic can be unit-tested and reused across booking/payment timers.

// DeadlineAfter returns now + hours. A non-positive hours yields now (immediate deadline).
func DeadlineAfter(now time.Time, hours int) time.Time {
	if hours <= 0 {
		return now
	}
	return now.Add(time.Duration(hours) * time.Hour)
}

// Overdue reports whether the deadline has passed at `now`.
func Overdue(deadline time.Time, now time.Time) bool {
	return !now.Before(deadline)
}

// HoursRemaining returns the whole hours remaining until deadline from now (0 if past).
func HoursRemaining(deadline time.Time, now time.Time) int {
	d := time.Until(deadline)
	if d <= 0 {
		return 0
	}
	return int(d.Hours())
}
