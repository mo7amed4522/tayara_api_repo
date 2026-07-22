package utils

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/mo7amed4522/tayara_api/plugin/email/global"

	"github.com/jordan-wright/email"
)

//@author: [mo7amed4522](https://github.com/mo7amed4522)
//@function: Email
//@description: Email sending method
//@param: subject string, body string
//@return: error

func Email(To, subject string, body string) error {
	to := strings.Split(To, ",")
	return send(to, subject, body)
}

//@author: [SliverHorn](https://github.com/SliverHorn)
//@function: ErrorToEmail
//@description: Send email to specified email address for email middleware error
//@param: subject string, body string
//@return: error

func ErrorToEmail(subject string, body string) error {
	to := strings.Split(global.GlobalConfig.To, ",")
	if to[len(to)-1] == "" { // Check if the last element of the slice is empty, if it is, remove it
		to = to[:len(to)-1]
	}
	return send(to, subject, body)
}

//@author: [mo7amed4522](https://github.com/mo7amed4522)
//@function: EmailTest
//@description: Email test method
//@param: subject string, body string
//@return: error

func EmailTest(subject string, body string) error {
	to := []string{global.GlobalConfig.To}
	return send(to, subject, body)
}

//@author: [mo7amed4522](https://github.com/mo7amed4522)
//@function: send
//@description: Email sending method
//@param: subject string, body string
//@return: error

func send(to []string, subject string, body string) error {
	from := global.GlobalConfig.From
	nickname := global.GlobalConfig.Nickname
	secret := global.GlobalConfig.Secret
	host := global.GlobalConfig.Host
	port := global.GlobalConfig.Port
	isSSL := global.GlobalConfig.IsSSL

	auth := smtp.PlainAuth("", from, secret, host)
	e := email.NewEmail()
	if nickname != "" {
		e.From = fmt.Sprintf("%s <%s>", nickname, from)
	} else {
		e.From = from
	}
	e.To = to
	e.Subject = subject
	e.HTML = []byte(body)
	var err error
	hostAddr := fmt.Sprintf("%s:%d", host, port)
	if isSSL {
		err = e.SendWithTLS(hostAddr, auth, &tls.Config{ServerName: host})
	} else {
		err = e.Send(hostAddr, auth)
	}
	return err
}
