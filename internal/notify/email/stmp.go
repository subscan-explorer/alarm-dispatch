package email

import (
	"context"
	"crypto/tls"

	"github.com/subscan-explorer/alarm-dispatch/conf"
	"github.com/subscan-explorer/alarm-dispatch/internal/model"
	"gopkg.in/gomail.v2"
)

type SMTP struct {
	cli  *gomail.Dialer
	from string
}

func (s *SMTP) SendEmail(_ context.Context, to []string, common map[string]string, alert model.Alert) error {
	msg := gomail.NewMessage()
	msg.SetHeader("From", s.from)
	msg.SetHeader("To", to...)
	msg.SetHeader("Subject", common["summary"])
	msg.SetBody("text/html", alert.HTML("<br>", "&nbsp;"))
	return s.cli.DialAndSend(msg)
}

func NewSMTP(c conf.Email) *SMTP {
	s := new(SMTP)
	s.cli = gomail.NewDialer(c.Host, c.Port, c.User, c.Secret)
	s.from = c.Sender
	s.cli.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	return s
}
