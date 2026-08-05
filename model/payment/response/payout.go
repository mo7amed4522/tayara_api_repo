package response

import (
	"time"

	paymentModel "github.com/mo7amed4522/tayara_api/model/payment"
)

// PayoutDetail a payout plus its funding transaction summary.
type PayoutDetail struct {
	paymentModel.Payout
	TransactionAmount int64      `json:"transactionAmount"`
	FeeAmount         int64      `json:"feeAmount"`
	CommissionAmount  int64      `json:"commissionAmount"`
	Currency          string     `json:"currency"`
	PaidAt            *time.Time `json:"paidAt"`
}

// AgentEarnings aggregates the agent's payout/earnings views (Agent Dashboard).
type AgentEarnings struct {
	AgentID             uint           `json:"agentId"`
	TotalEarned         int64          `json:"totalEarned"`         // sum of captured agent shares
	TotalPendingPayouts int64          `json:"totalPendingPayouts"` // sum of pending payouts
	TotalPaidOut        int64          `json:"totalPaidOut"`        // sum of paid payouts
	PendingPayouts      []PayoutDetail `json:"pendingPayouts"`
	RecentPayouts       []PayoutDetail `json:"recentPayouts"`
}
