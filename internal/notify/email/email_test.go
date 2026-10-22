package email

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/subscan-explorer/alarm-dispatch/internal/model"
)

type mockSender struct {
	err        error
	ctx        context.Context
	recipients []string
	alert      model.Alert
}

// SendEmail records delivery parameters without contacting an email provider.
func (s *mockSender) SendEmail(ctx context.Context, recipients []string, alert model.Alert) error {
	s.ctx, s.recipients, s.alert = ctx, recipients, alert
	return s.err
}

func TestNotifyUsesSenderAndPropagatesFailure(t *testing.T) {
	for _, fail := range []bool{false, true} {
		t.Run(map[bool]string{false: "success", true: "failure"}[fail], func(t *testing.T) {
			sender := &mockSender{}
			if fail {
				sender.err = errors.New("mock delivery failure")
			}
			notifier := &Email{sender: sender}
			notifier.conf.To = []string{"local@example.invalid"}
			ctx := context.Background()
			alert := model.Alert{Status: model.AlertStatusFiring, Labels: map[string]string{"team": "ops"}}
			retry, err := notifier.Notify(ctx, alert)
			if retry != fail || !errors.Is(err, sender.err) {
				t.Fatalf("retry=%v, err=%v; want retry=%v, err=%v", retry, err, fail, sender.err)
			}
			if sender.ctx != ctx || !reflect.DeepEqual(sender.recipients, notifier.conf.To) || !reflect.DeepEqual(sender.alert, alert) {
				t.Fatal("email context, recipients or alert changed before delivery")
			}
		})
	}
}
