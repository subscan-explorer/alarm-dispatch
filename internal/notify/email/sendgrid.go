package email

import (
	"context"
	"errors"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/subscan-explorer/alarm-dispatch/conf"
	"github.com/subscan-explorer/alarm-dispatch/internal/model"
)

type Sendgrid struct {
	cli  *sendgrid.Client
	from string
}

func (s *Sendgrid) SendEmail(_ context.Context, to []string, common map[string]string, alert model.Alert) error {
	for _, rc := range to {
		message := mail.NewSingleEmail(
			mail.NewEmail("Subscan", s.from),
			alert.GetTitle(),
			mail.NewEmail("", rc),
			"",
			alert.HTML("<br>", "&nbsp;"))
		rsp, err := s.cli.Send(message)
		if err != nil {
			return err
		}
		if rsp.StatusCode > 299 {
			return errors.New(rsp.Body)
		}
	}
	return nil
}

func NewSendgrid(c conf.Email) *Sendgrid {
	s := new(Sendgrid)
	s.cli = sendgrid.NewSendClient(c.Secret)
	s.from = c.Sender
	return s
}
