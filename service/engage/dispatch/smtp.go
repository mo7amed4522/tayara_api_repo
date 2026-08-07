package dispatch

import (
	"net/smtp"
	"strconv"

	"github.com/jordan-wright/email"
	"github.com/mo7amed4522/tayara_api/global"
)

// SmtpSender delivers email notifications via SMTP using the jordan-wright/email library.
// The config comes from `email.*` (host/port/SSL/from/secret/nickname). Delivery is best-effort:
// a failure is logged but never fails the calling flow.
type SmtpSender struct{}

// Enabled reports whether the SMTP config is populated enough to send.
func (SmtpSender) Enabled() bool {
	cfg := global.GVA_CONFIG.Email
	return cfg.Host != "" && cfg.From != "" && cfg.Secret != "" && cfg.Port != 0
}

// Send composes and sends a plain-text (UTF-8) email to one or more recipients.
func (SmtpSender) Send(to []string, subject, body string) error {
	cfg := global.GVA_CONFIG.Email
	if len(to) == 0 || !smtpEnabled() {
		return nil
	}
	m := email.NewEmail()
	if cfg.Nickname != "" {
		m.From = cfg.Nickname + " <" + cfg.From + ">"
	} else {
		m.From = cfg.From
	}
	m.To = to
	m.Subject = subject
	m.Text = []byte(body)
	addr := cfg.Host + ":" + strconv.Itoa(cfg.Port)
	auth := smtp.PlainAuth("", cfg.From, cfg.Secret, cfg.Host)
	if cfg.IsSSL {
		return m.SendWithTLS(addr, auth, nil)
	}
	return m.Send(addr, auth)
}

func smtpEnabled() bool {
	cfg := global.GVA_CONFIG.Email
	return cfg.Host != "" && cfg.From != "" && cfg.Secret != "" && cfg.Port != 0
}
