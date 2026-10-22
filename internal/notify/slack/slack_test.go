package slack

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/subscan-explorer/alarm-dispatch/internal/model"
)

func TestWebhookFailureAndCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()
	notifier := &Notifier{Webhook: server.URL}
	alert := model.Alert{Status: model.AlertStatusFiring}
	if retry, err := notifier.Notify(context.Background(), alert); err == nil || retry {
		t.Fatalf("HTTP rejection: retry=%v, err=%v; want non-retryable error", retry, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if retry, err := notifier.Notify(ctx, alert); !errors.Is(err, context.Canceled) || !retry {
		t.Fatalf("canceled transport: retry=%v, err=%v", retry, err)
	}
}
