package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

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
		return true, err
	}
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	if rsp, err = cli.HTTPCli.Do(req); err != nil {
		return true, err
	}
	defer rsp.Body.Close()
	_, _ = io.Copy(io.Discard, rsp.Body)
	if rsp.StatusCode != 200 {
		return false, errors.New(rsp.Status)
	}
	return false, nil
}

func (n *Notifier) buildMessage(_ map[string]string, alert model.Alert) Message {
	msg := Message{}
	msg.Blocks = append(msg.Blocks, Block{
		Type: "header",
		Text: &Text{
			Type: "plain_text",
			Text: alert.Status,
		},
	})
	tmBlock := newSectionBlock()
	tmBlock.Fields = append(tmBlock.Fields,
		newMdText(fmt.Sprintf("*Start:* \n%s", alert.StartsAt.Format(time.RFC3339))))
	if !alert.EndsAt.IsZero() {
		tmBlock.Fields = append(tmBlock.Fields,
			newMdText(fmt.Sprintf("*End:* \n%s", alert.EndsAt.Format(time.RFC3339))))
	}
	msg.Blocks = append(msg.Blocks, tmBlock)
	for k, v := range alert.Annotations {
		if len(k) == 0 || len(v) == 0 {
			continue
		}
		contentBlock := newSectionBlock()
		contentBlock.Text = newMdText(fmt.Sprintf("*%s:* \n%s", strings.Title(k), v))
		msg.Blocks = append(msg.Blocks, contentBlock)
	}

	if len(alert.Labels) != 0 {
		buf := model.GetByteBuf()
		for k, v := range alert.Labels {
			buf.WriteString(" • ")
			buf.WriteByte('`')
			buf.WriteString(k)
			buf.WriteString("`: `")
			buf.WriteString(v)
			buf.WriteByte('`')
			buf.WriteByte('\n')
		}
		labelBlock := newSectionBlock()
		labelBlock.Text = newMdText(fmt.Sprintf("*Tag:* \n%s", buf.String()))
		msg.Blocks = append(msg.Blocks, labelBlock)
		model.PutByteBuf(buf)
	}

	return msg
}

type Message struct {
	Blocks []Block `json:"blocks,omitempty"`
}

type Block struct {
	Type   string  `json:"type,omitempty"`
	Fields []*Text `json:"fields,omitempty"`
	Text   *Text   `json:"text,omitempty"`
}

func newSectionBlock() Block {
	return Block{
		Type: "section",
	}
}

type Text struct {
	Type string `json:"type,omitempty"`
	Text string `json:"text,omitempty"`
}

func newMdText(msg string) *Text {
	return &Text{
		Type: "mrkdwn",
		Text: msg,
	}
}
