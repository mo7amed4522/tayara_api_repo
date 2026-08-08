package export

import (
	"bytes"
	"encoding/csv"
	"errors"
	"strconv"

	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/model/booking"
	"github.com/mo7amed4522/tayara_api/model/bxapi"
	"github.com/mo7amed4522/tayara_api/model/dxapi"
	"github.com/mo7amed4522/tayara_api/model/payment"
)

// ExportServices produces CSV bytes for admin data export. Money is written in minor units
// (AED fils); consumers divide where a currency string is expected.
type ExportServices struct{}

// csvFromRows renders a header + rows into a UTF-8 CSV byte slice.
func csvFromRows(header []string, rows [][]string) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	if err := w.Write(header); err != nil {
		return nil, err
	}
	for _, r := range rows {
		if err := w.Write(r); err != nil {
			return nil, err
		}
	}
	w.Flush()
	return buf.Bytes(), w.Error()
}

// Users exports end-user profiles (admin; optional status filter).
func (s ExportServices) Users(status string) ([]byte, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	db := global.GVA_DB.Model(&bxapi.BxsUserProfile{})
	if status != "" {
		db = db.Where("status = ?", status)
	}
	var list []bxapi.BxsUserProfile
	if err := db.Order("created_at desc").Find(&list).Error; err != nil {
		return nil, err
	}
	rows := make([][]string, 0, len(list))
	for _, u := range list {
		rows = append(rows, []string{
			strconv.FormatUint(uint64(u.ID), 10), u.FullName, u.Username, u.Email,
			u.PhoneCountryCode + u.Phone, u.Status, u.KYCStatus,
		})
	}
	return csvFromRows([]string{"id", "full_name", "username", "email", "phone", "status", "kyc_status"}, rows)
}

// Agents generates agent profiles (admin; optional status filter).
func (s ExportServices) Agents(status string) ([]byte, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	db := global.GVA_DB.Model(&dxapi.DxAgentProfile{})
	if status != "" {
		db = db.Where("status = ?", status)
	}
	var list []dxapi.DxAgentProfile
	if err := db.Find(&list).Error; err != nil {
		return nil, err
	}
	rows := make([][]string, 0, len(list))
	for _, a := range list {
		rows = append(rows, []string{
			strconv.FormatUint(uint64(a.ID), 10), a.CompanyName, a.ResponsibleName,
			a.BusinessEmail, a.PhoneCountryCode + a.Phone, a.Country, a.LicenseNumber, a.Status,
		})
	}
	return csvFromRows([]string{"id", "company_name", "responsible_name", "business_email", "phone", "country", "license_number", "status"}, rows)
}

// Bookings generates bookings (admin; optional status filter).
func (s ExportServices) Bookings(status string) ([]byte, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	db := global.GVA_DB.Model(&booking.Booking{})
	if status != "" {
		db = db.Where("booking_status = ?", status)
	}
	var list []booking.Booking
	if err := db.Order("created_at desc").Find(&list).Error; err != nil {
		return nil, err
	}
	rows := make([][]string, 0, len(list))
	for _, b := range list {
		rows = append(rows, []string{
			strconv.FormatUint(uint64(b.ID), 10), b.BookingCode,
			strconv.FormatUint(uint64(b.UserID), 10), strconv.FormatUint(uint64(b.AgentID), 10),
			b.BookingStatus, b.PaymentStatus, strconv.FormatInt(b.AmountGross, 10),
		})
	}
	return csvFromRows([]string{"id", "booking_code", "user_id", "agent_id", "booking_status", "payment_status", "amount_gross"}, rows)
}

// Payments generates payment transactions (admin; optional status filter).
func (s ExportServices) Payments(status string) ([]byte, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	db := global.GVA_DB.Model(&payment.PaymentTransaction{})
	if status != "" {
		db = db.Where("status = ?", status)
	}
	var list []payment.PaymentTransaction
	if err := db.Order("created_at desc").Find(&list).Error; err != nil {
		return nil, err
	}
	rows := make([][]string, 0, len(list))
	for _, p := range list {
		rows = append(rows, []string{
			strconv.FormatUint(uint64(p.ID), 10),
			strconv.FormatUint(uint64(p.BookingID), 10),
			strconv.FormatUint(uint64(p.AgentID), 10),
			p.PaymentMethod, p.Currency, strconv.FormatInt(p.Amount, 10),
			strconv.FormatInt(p.FeeAmount, 10), strconv.FormatInt(p.CommissionAmount, 10),
			strconv.FormatInt(p.AgentShare, 10), p.Status,
		})
	}
	return csvFromRows([]string{"id", "booking_id", "agent_id", "method", "currency", "amount", "fee_amount", "commission_amount", "agent_share", "status"}, rows)
}