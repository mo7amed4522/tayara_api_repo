package service

import (
	"github.com/mo7amed4522/tayara_api/plugin/email/utils"
)

type EmailService struct{}

//@author: [mo7amed4522](https://github.com/mo7amed4522)
//@function: EmailTest
//@description: Send test email
//@return: err error

func (e *EmailService) EmailTest() (err error) {
	subject := "test"
	body := "test"
	err = utils.EmailTest(subject, body)
	return err
}

//@author: [mo7amed4522](https://github.com/mo7amed4522)
//@function: EmailTest
//@description: Send test email
//@return: err error
//@params to string 	 Recipient
//@params subject string   Subject
//@params body  string   Email content

func (e *EmailService) SendEmail(to, subject, body string) (err error) {
	err = utils.Email(to, subject, body)
	return err
}
