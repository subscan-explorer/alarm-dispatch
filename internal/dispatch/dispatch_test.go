package dispatch

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/subscan-explorer/alarm-dispatch/conf"
	"github.com/subscan-explorer/alarm-dispatch/internal/model"
	"github.com/subscan-explorer/alarm-dispatch/internal/notify"
)

func TestDispatchFiltersRoutesAndSendsLocalWebhook(t *testing.T) {
	var mu sync.Mutex
	var payloads []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
			t.Errorf("unexpected webhook request: %s, %s", r.Method, r.Header.Get("Content-Type"))
		}
		var body json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode webhook: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		mu.Lock()
		payloads = append(payloads, string(body))
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	originalConf := conf.Conf
	conf.Conf = conf.Config{
		Alert: conf.Alert{Filter: conf.Filter{Label: conf.LabelKV{Value: []string{"^ignore$"}}}},
		Dispatch: conf.Dispatch{LabelMatch: []conf.LabelMatch{{
			Receiver: []string{"local", "local"},
			LabelKV:  conf.LabelKV{Combination: []map[string]string{{"^team$": "^ops$"}}},
		}}},
		Receivers: []conf.Receiver{{Name: "local", Type: "slack", Webhook: server.URL}},
	}
	conf.Conf.Label.Exclude.Label.Key = []string{"^private$"}
	p, one = nil, sync.Once{}
	notify.Init()
	t.Cleanup(func() {
		conf.Conf = conf.Config{}
		notify.Init()
		conf.Conf = originalConf
		p, one = nil, sync.Once{}
	})

	started := time.Date(2026, time.October, 22, 0, 0, 0, 0, time.UTC)
	Dispatch(model.Notification{Alerts: []model.Alert{
		{Status: model.AlertStatusFiring, Labels: map[string]string{"team": "ops", "severity": "ignore"}},
		{
			Status:      model.AlertStatusFiring,
			Labels:      map[string]string{"team": "ops", "private": "hidden"},
			Annotations: map[string]string{"summary": "local regression"},
			StartsAt:    started,
		},
		{
			Status:   model.AlertStatusResolved,
			Labels:   map[string]string{"team": "ops"},
			StartsAt: started,
			EndsAt:   started.Add(time.Minute),
		},
	}})

	mu.Lock()
	defer mu.Unlock()
	if len(payloads) != 2 {
		t.Fatalf("got %d notifications, want one firing and one resolved", len(payloads))
	}
	combined := strings.Join(payloads, "\n")
	for _, want := range []string{"firing", "resolved", "local regression", "2026-10-22T00:00:00Z", "2026-10-22T00:01:00Z"} {
		if !strings.Contains(combined, want) {
			t.Errorf("notification payload missing %q: %s", want, combined)
		}
	}
	for _, excluded := range []string{"ignore", "private", "hidden"} {
		if strings.Contains(combined, excluded) {
			t.Errorf("filtered value %q reached webhook: %s", excluded, combined)
		}
	}
}
