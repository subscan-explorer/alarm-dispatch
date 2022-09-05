package email

import (
	"context"
	"log"
	"strings"
	"sync"

	"github.com/subscan-explorer/alarm-dispatch/conf"
	"github.com/subscan-explorer/alarm-dispatch/internal/model"
)

type Sender interface {
	SendEmail(context.Context, []string, map[string]string, model.Alert) error
}

// Email implements a Notifier for email notifications.
type Email struct {
	conf struct {
		To []string
	}
	sender Sender
}

var once sync.Once
var sender Sender

// New returns a new Email notifier.
func New(c conf.Receiver) *Email {
	once.Do(initSender)
	e := new(Email)
	e.conf.To = c.Email
	e.sender = sender
	return e
}

// Notify implements the Notifier interface.
func (e *Email) Notify(ctx context.Context, common map[string]string, alert model.Alert) (bool, error) {
	return false, e.sender.SendEmail(ctx, e.conf.To, common, alert)
}

func initSender() {
	switch strings.ToLower(conf.Conf.Email.Type) {
	case "sendgrid":
		sender = NewSendgrid(conf.Conf.Email)
		return
	case "smtp":
		sender = NewSMTP(conf.Conf.Email)
		return
	}
	log.Fatalln("not configure email send account")
}
