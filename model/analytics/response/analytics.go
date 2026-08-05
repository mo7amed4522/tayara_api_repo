package response

// AdminUsersCard aggregates the Admin Dashboard Users card.
type AdminUsersCard struct {
	TotalUsers       int64 `json:"totalUsers"`
	NewUsersThisWeek int64 `json:"newUsersThisWeek"`
	TotalSpend       int64 `json:"totalSpend"` // minor units
	KYCApproved      int64 `json:"kycApproved"`
	ActiveBookings   int64 `json:"activeBookings"` // non-terminal bookings
	Shares           int64 `json:"shares"`
	Reviews          int64 `json:"reviews"`
}

// AdminAgentsCard aggregates the Admin Dashboard Agents card.
type AdminAgentsCard struct {
	TotalAgents   int64 `json:"totalAgents"`
	Pending       int64 `json:"pending"`
	Active        int64 `json:"active"`
	Suspended     int64 `json:"suspended"`
	Rejected      int64 `json:"rejected"`
	TotalPackages int64 `json:"totalPackages"` // active listings across agents
}

// AdminBusinessCard aggregates the Admin Dashboard Business card (money = minor units).
type AdminBusinessCard struct {
	GMV        int64 `json:"gmv"`        // gross value of all bookings
	Commission int64 `json:"commission"` // platform commission on captured payments
	Fees       int64 `json:"fees"`       // platform stripe fees on captured payments
	Payouts    int64 `json:"payouts"`    // total paid out to agents
	Refunds    int64 `json:"refunds"`    // total refunded
	NetProfit  int64 `json:"netProfit"`  // commission + fees - refunds
}

// AgentCard aggregates one agent's dashboard metrics (money = minor units).
type AgentCard struct {
	AgentID              uint    `json:"agentId"`
	TotalBookings        int64   `json:"totalBookings"`
	ActiveBookings       int64   `json:"activeBookings"`
	CompletedBookings    int64   `json:"completedBookings"`
	Revenue              int64   `json:"revenue"` // sum of captured agent shares
	PendingPayouts       int64   `json:"pendingPayouts"`
	TotalPackages        int64   `json:"totalPackages"`
	PackageViews         int64   `json:"packageViews"`
	AvgConfirmationHours float64 `json:"avgConfirmationHours"` // from requested -> price_confirmed
	SLAEnabled           bool    `json:"slaEnabled"`
	SLAHours             int     `json:"slaHours"`
}
