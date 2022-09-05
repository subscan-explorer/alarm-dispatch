package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/subscan-explorer/alarm-dispatch/conf"
	"github.com/subscan-explorer/alarm-dispatch/internal/model"
	"github.com/subscan-explorer/alarm-dispatch/internal/notify/cli"
)

// Notifier implements a Notifier for Slack notifications.
type Notifier struct {
	conf struct {
		Webhook string
	}
}

// New returns a new Slack notification handler.
func New(c conf.Receiver) *Notifier {
	return &Notifier{
		conf: struct{ Webhook string }{
			Webhook: c.Webhook,
		},
	}
}

// Notify implements the Notifier interface.
func (n *Notifier) Notify(ctx context.Context, common map[string]string, alert model.Alert) (bool, error) {
	if len(common) == 0 {
		return false, nil
	}
	var (
		reqBody, _ = json.Marshal(n.buildMessage(common, alert))
		req        *http.Request
		rsp        *http.Response
		err        error
	)
	if req, err = http.NewRequestWithContext(ctx, http.MethodPost, n.conf.Webhook, bytes.NewReader(reqBody)); err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	if rsp, err = cli.HTTPCli.Do(req); err != nil {
		return false, err
	}
	defer rsp.Body.Close()
	_, _ = io.Copy(io.Discard, rsp.Body)
	if rsp.StatusCode != 200 {
		return false, errors.New(rsp.Status)
	}
	return false, nil
}

func (n *Notifier) buildMessage(common map[string]string, alerts model.Alert) Message {
	msg := Message{}
	msg.Blocks = append(msg.Blocks, Block{
		Type: "header",
		Text: Text{
			Type: "plain_text",
			Text: common["summary"],
		},
	})
	msg.Blocks = append(msg.Blocks, Block{
		Type: "section",
		Text: Text{
			Type: "mrkdwn",
			Text: alerts.Markdown(),
		},
	})
	return msg
}

type Message struct {
	Blocks []Block `json:"blocks"`
}

type Block struct {
	Type string `json:"type"`
	Text Text   `json:"text"`
}

type Text struct {
	Type string `json:"type"`
	Text string `json:"text"`
}
