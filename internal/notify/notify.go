package notify

import (
	"context"
	"strings"

	"github.com/subscan-explorer/alarm-dispatch/conf"
	"github.com/subscan-explorer/alarm-dispatch/internal/model"
	"github.com/subscan-explorer/alarm-dispatch/internal/notify/email"
	"github.com/subscan-explorer/alarm-dispatch/internal/notify/matrix"
	"github.com/subscan-explorer/alarm-dispatch/internal/notify/slack"
	"github.com/subscan-explorer/alarm-dispatch/internal/notify/telegram"
)

var noticer map[string]Notifier

type Notifier interface {
	Notify(context.Context, model.Alert) (bool, error)
	RemoveLastMessage(context.Context)
}

func Init() {
	noticer = make(map[string]Notifier)
	for _, receiver := range conf.Conf.Receivers {
		var rc Notifier
		switch strings.ToLower(receiver.Type) {
		case "slack":
			rc = slack.New(receiver)
		case "telegram":
			rc = telegram.New(receiver)
		case "email":
			rc = email.New(receiver)
		case "element":
			rc = matrix.New(receiver)
		}
		if rc != nil {
			noticer[receiver.Name] = rc
		}
	}
}

func GetNoticer(name string) Notifier {
	return noticer[name]
}
