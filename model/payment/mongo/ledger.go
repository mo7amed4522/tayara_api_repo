package mongo

import "time"

// LedgerEvent an append-only payment lifecycle event — the operational record of truth.
// The Mongo ledger is where every state transition is logged first; the authoritative
// MySQL row is written alongside. Event types: created|authorized|captured|refunded|failed.
type LedgerEvent struct {
	EventID   string    `bson:"event_id"`
	PaymentID string    `bson:"payment_id"`
	Type      string    `bson:"type"`
	Provider  string    `bson:"provider"`
	Method    string    `bson:"method"`
	Amount    int64     `bson:"amount"`
	FeeAmount int64     `bson:"fee_amount"`
	Currency  string    `bson:"currency"`
	Actor     string    `bson:"actor"` // system|admin|user
	Note      string    `bson:"note"`
	At        time.Time `bson:"at"`
}

// PaymentState the current payment lifecycle state (one doc per payment, idempotent). Reads and
// writes are guarded by the lifecycle red lock so no two mutations can race.
type PaymentState struct {
	PaymentID string    `bson:"_id"`
	BookingID uint      `bson:"booking_id"`
	Provider  string    `bson:"provider"`
	Method    string    `bson:"method"`
	Status    string    `bson:"status"`
	Amount    int64     `bson:"amount"`
	FeeAmount int64     `bson:"fee_amount"`
	Currency  string    `bson:"currency"`
	UpdatedAt time.Time `bson:"updated_at"`
}
