package response

// SearchHit is a compact, cross-entity search result. The admin global search fans a single
// keyword across users, agents, packages and bookings and returns a small list per group.
type SearchResult struct {
	Keyword  string       `json:"keyword"`
	Users    []UserHit    `json:"users"`
	Agents   []AgentHit   `json:"agents"`
	Packages []PackageHit `json:"packages"`
	Bookings []BookingHit `json:"bookings"`
	Total    int64        `json:"total"`
}

// UserHit a matching end-user (bx) profile.
type UserHit struct {
	ID     uint   `json:"id"`
	Name   string `json:"name"` // fullName (falls back to nickname/username)
	Email  string `json:"email"`
	Phone  string `json:"phone"`
	Status string `json:"status"`
}

// AgentHit a matching travel agent (dx) profile.
type AgentHit struct {
	ID            uint   `json:"id"`
	CompanyName   string `json:"companyName"`
	BusinessEmail string `json:"businessEmail"`
	Status        string `json:"status"`
}

// PackageHit a matching catalog package.
type PackageHit struct {
	ID          uint   `json:"id"`
	AgentID     uint   `json:"agentId"`
	Name        string `json:"name"`
	Status      string `json:"status"`
	DisplayPrice int64 `json:"displayPrice"` // minor units
}

// BookingHit a matching booking.
type BookingHit struct {
	ID           uint   `json:"id"`
	BookingCode  string `json:"bookingCode"`
	BookingStatus string `json:"bookingStatus"`
	AmountGross  int64  `json:"amountGross"` // minor units
}