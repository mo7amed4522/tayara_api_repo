package request

// TravelerDocReq one document entry submitted for a traveler.
type TravelerDocReq struct {
	TravelerID     uint   `json:"travelerId"`
	DocumentTypeID uint   `json:"documentTypeId"`
	FileURL        string `json:"fileUrl"`
	OriginalName   string `json:"originalName"`
	MIME           string `json:"mime"`
}

// SubmitDocumentsReq user submits (or resubmits after correction) traveler documents.
type SubmitDocumentsReq struct {
	BookingID uint             `json:"bookingId" binding:"required"`
	Documents []TravelerDocReq `json:"documents" binding:"required"`
}

// ApproveDocumentReq agent approves a traveler document or requests a correction.
type ApproveDocumentReq struct {
	DocumentID uint   `json:"documentId" binding:"required"`
	Approve    bool   `json:"approve"` // true = approve, false = request correction
	Note       string `json:"note"`
}

// BookingDocumentReq upload a booking-level document (agent confirmation, issued ticket, KYC file).
type BookingDocumentReq struct {
	BookingID    uint   `json:"bookingId" binding:"required"`
	DocRole      string `json:"docRole" binding:"required"` // agent_confirmation|issued_ticket|kyc_file
	FileURL      string `json:"fileUrl" binding:"required"`
	OriginalName string `json:"originalName"`
}

// KYCReviewReq admin approves or rejects the user KYC for a booking.
type KYCReviewReq struct {
	BookingID uint   `json:"bookingId" binding:"required"`
	Approve   bool   `json:"approve"` // true = approved, false = rejected
	Note      string `json:"note"`
}
