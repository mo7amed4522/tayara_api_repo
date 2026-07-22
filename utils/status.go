package utils

// Define status bits
const (
	InitialStatus        int8 = 0b0000 // initial state
	ImmediatePushSuccess int8 = 0b0001 // The status indicates a successful push notification. If the push notification fails, the status remains the initial state, but it will not be confused with the status of push notifications the following day. Instead, it can be described as an overlapping status.
	SecondDayPushSuccess int8 = 0b0010 // The status indicates a successful push notification on the second day.
	SecondDayPushFail    int8 = 0b0100 // The status indicates a failed push notification on the second day.
)

// SetStatus Set status
func SetStatus(currentStatus int8, statusToSet int8) int8 {
	return currentStatus | statusToSet
}

// ClearStatus Clear status
func ClearStatus(currentStatus int8, statusToClear int8) int8 {
	return currentStatus &^ statusToClear
}

// CheckStatus Check status
func CheckStatus(currentStatus int8, statusToCheck int8) bool {
	return currentStatus&statusToCheck != 0
}
