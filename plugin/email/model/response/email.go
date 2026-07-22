package response

type Email struct {
	To      string `json:"to"`      // Email sent to who
	Subject string `json:"subject"` // Email subject
	Body    string `json:"body"`    // Email content
}
