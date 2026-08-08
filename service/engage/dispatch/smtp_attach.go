package dispatch

import (
	"bytes"
	"net/smtp"
	"strconv"

	"github.com/jordan-wright/email"
	"github.com/mo7amed4522/tayara_api/global"
)

// SendWithAttachment composes an email with a single PDF attachment (used for e-tickets).
// The PDF bytes are attached inline as an application/pdf part addressed to the recipient.
func (SmtpSender) SendWithAttachment(to []string, subject, body, filename string, pdfBytes []byte) error {
	if len(to) == 0 || !smtpEnabled() || len(pdfBytes) == 0 {
		return nil
	}
	cfg := global.GVA_CONFIG.Email
	m := email.NewEmail()
	if cfg.Nickname != "" {
		m.From = cfg.Nickname + " <" + cfg.From + ">"
	} else {
		m.From = cfg.From
	}
	m.To = to
	m.Subject = subject
	m.Text = []byte(body)
	m.Attach(bytes.NewReader(pdfBytes), filename, "application/pdf")
	addr := cfg.Host + ":" + strconv.Itoa(cfg.Port)
	auth := smtp.PlainAuth("", cfg.From, cfg.Secret, cfg.Host)
	if cfg.IsSSL {
		return m.SendWithTLS(addr, auth, nil)
	}
	return m.Send(addr, auth)
}
