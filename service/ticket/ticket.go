package ticket

import (
	"bytes"
	"errors"
	"strconv"
	"time"

	"github.com/go-pdf/fpdf"
	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/model/booking"
	"github.com/mo7amed4522/tayara_api/model/bxapi"
	"github.com/mo7amed4522/tayara_api/model/catalog"
	"github.com/mo7amed4522/tayara_api/model/dxapi"
	"github.com/mo7amed4522/tayara_api/service/engage/dispatch"
	"go.uber.org/zap"
)

// TicketServices generates and distributes booking e-tickets.
type TicketServices struct{}

// ticketData carries the resolved fields rendered onto the e-ticket PDF.
type ticketData struct {
	BookingCode  string
	Status       string
	Amount       string
	TravelStart  string
	TravelEnd    string
	Travelers    int
	UserFullName string
	UserEmail    string
	AgentCompany string
	PackageName  string
}

// resolve loads the booking + buyer + agent/package context for a ticket.
func (t TicketServices) resolve(bookingID uint) (*ticketData, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	var bk booking.Booking
	if err := global.GVA_DB.First(&bk, bookingID).Error; err != nil {
		return nil, err
	}
	d := &ticketData{
		BookingCode: bk.BookingCode,
		Status:      bk.BookingStatus,
		Travelers:   bk.TravelerCount,
		Amount:      money(bk.AmountGross, bk.Currency),
	}
	if bk.StartDate != nil {
		d.TravelStart = bk.StartDate.Format("2006-01-02")
	}
	if bk.EndDate != nil {
		d.TravelEnd = bk.EndDate.Format("2006-01-02")
	}
	var user bxapi.BxsUserProfile
	if err := global.GVA_DB.First(&user, bk.UserID).Error; err == nil {
		d.UserFullName = user.FullName
		d.UserEmail = user.Email
	}
	var agent dxapi.DxAgentProfile
	if err := global.GVA_DB.First(&agent, bk.AgentID).Error; err == nil {
		d.AgentCompany = agent.CompanyName
	}
	if bk.PackageID != nil {
		var pkg catalog.Package
		if err := global.GVA_DB.First(&pkg, *bk.PackageID).Error; err == nil {
			d.PackageName = pkg.Name
		}
	}
	return d, nil
}

// GenerateBookingTicket builds a single-page A4 PDF e-ticket for a booking and returns its bytes.
func (t TicketServices) GenerateBookingTicket(bookingID uint) ([]byte, error) {
	d, err := t.resolve(bookingID)
	if err != nil {
		return nil, err
	}
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.AddPage()

	pdf.SetFont("Helvetica", "B", 18)
	pdf.Cell(0, 10, "Tayara e-Ticket")
	pdf.Ln(14)

	pdf.SetFont("Courier", "B", 12)
	pdf.Cell(0, 8, "BOOKING "+d.BookingCode)
	pdf.Ln(3)
	pdf.SetFont("Helvetica", "", 10)
	pdf.SetTextColor(80, 80, 80)
	pdf.Cell(0, 8, "Status: "+d.Status)
	pdf.SetTextColor(0, 0, 0)
	pdf.Ln(12)

	pdf.SetFont("Helvetica", "", 10)
	label("Traveler", d.UserFullName, pdf)
	label("Email", d.UserEmail, pdf)
	label("Agent", d.AgentCompany, pdf)
	label("Package", d.PackageName, pdf)
	label("Travel dates", dateRange(d.TravelStart, d.TravelEnd), pdf)
	label("Travelers", strconv.Itoa(d.Travelers), pdf)
	label("Amount", d.Amount, pdf)
	pdf.Ln(6)
	pdf.SetFont("Helvetica", "I", 8)
	pdf.SetTextColor(120, 120, 120)
	pdf.MultiCell(0, 5, "Generated "+time.Now().UTC().Format(time.RFC3339), "", "L", false)

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// SendBookingTicket generates the ticket PDF and emails it to the buyer via SMTP. Best-effort:
// if SMTP is disabled the ticket is not emailed but the method reports it so the caller can
// surface the status instead of erroring the flow.
func (t TicketServices) SendBookingTicket(adminID, bookingID uint) (string, error) {
	d, err := t.resolve(bookingID)
	if err != nil {
		return "", err
	}
	if d.UserEmail == "" {
		return "", errors.New("booking buyer has no email to send the ticket to")
	}
	pdfBytes, err := t.GenerateBookingTicket(bookingID)
	if err != nil {
		return "", err
	}
	var smtp dispatch.SmtpSender
	if !smtp.Enabled() {
		return "smtp unavailable", nil
	}
	if err := smtp.SendWithAttachment([]string{d.UserEmail},
		"Your Tayara e-Ticket "+d.BookingCode,
		"Please find attached the e-ticket for booking "+d.BookingCode+".", d.BookingCode+".pdf", pdfBytes); err != nil {
		global.GVA_LOG.Warn("ticket email failed", zap.Error(err))
		return "smtp failed", nil
	}
	return "sent", nil
}

// money formats minor-unit (fils) money into "1,234.00 AED".
func money(amount int64, currency string) string {
	if currency == "" {
		currency = "AED"
	}
	// split fils/whole units
	whole := amount / 100
	fils := amount % 100
	num := strconv.FormatInt(whole, 10)
	// thousands separator
	for i := len(num) - 3; i > 0; i -= 3 {
		num = num[:i] + "," + num[i:]
	}
	return num + "." + fmtTwo(fils) + " " + currency
}

// fmtTwo renders a two-digit fil fraction (e.g. 5 -> "05").
func fmtTwo(n int64) string {
	if n < 10 {
		return "0" + strconv.FormatInt(n, 10)
	}
	return strconv.FormatInt(n, 10)
}

// label prints a left key and right value on one row of the ticket.
func label(key, value string, pdf *fpdf.Fpdf) {
	pdf.SetFont("Helvetica", "B", 10)
	pdf.Cell(60, 8, key+":")
	pdf.SetFont("Helvetica", "", 10)
	pdf.Cell(0, 8, value)
	pdf.Ln(8)
}

// dateRange returns a single or spanning date label.
func dateRange(start, end string) string {
	if start == "" && end == "" {
		return "—"
	}
	if start == "" {
		return end
	}
	if end == "" {
		return start
	}
	return start + "  →  " + end
}