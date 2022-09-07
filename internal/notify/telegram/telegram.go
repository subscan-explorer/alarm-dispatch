package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/subscan-explorer/alarm-dispatch/conf"
	"github.com/subscan-explorer/alarm-dispatch/internal/model"
	"github.com/subscan-explorer/alarm-dispatch/internal/notify/cli"
)

// Notifier implements a Notifier for Telegram notifications.
type Notifier struct {
	conf struct {
		Webhook string
		ChatID  string
	}
}

// New returns a new Telegram notification handler.
func New(c conf.Receiver) *Notifier {
	const host = "https://api.telegram.org/bot"
	notify := new(Notifier)
	if len(c.Token) != 0 {
		notify.conf.Webhook = fmt.Sprintf("%s%s/", host, c.Token)
	} else {
		notify.conf.Webhook = c.Webhook
	}
	notify.conf.ChatID = c.ChatID
	return notify
}

func (n *Notifier) Notify(ctx context.Context, common map[string]string, alert model.Alert) (bool, error) {
	var (
		reqBody, _ = json.Marshal(n.buildMessage(common, alert))
		req        *http.Request
		host       = n.conf.Webhook + "sendMessage"
		rsp        *http.Response
		err        error
	)
	fmt.Println(string(reqBody))
	if req, err = http.NewRequestWithContext(ctx, http.MethodPost, host, bytes.NewReader(reqBody)); err != nil {
		return true, err
	}
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	if rsp, err = cli.HTTPCli.Do(req); err != nil {
		return true, err
	}
	defer rsp.Body.Close()
	//_, _ = io.Copy(io.Discard, rsp.Body)
	rspBody, _ := io.ReadAll(rsp.Body)
	fmt.Println(string(rspBody))
	if rsp.StatusCode != 200 {
		return false, errors.New(rsp.Status)
	}
	return false, nil
}

func (n *Notifier) buildMessage(_ map[string]string, alert model.Alert) Message {
	return Message{
		ChatID:    n.conf.ChatID,
		Text:      alert.HTML("\n", ""),
		ParseMode: "Html",
	}
}

type Message struct {
	ChatID    string `json:"chat_id,omitempty"`
	Text      string `json:"text,omitempty"`
	ParseMode string `json:"parse_mode,omitempty"` // MarkdownV2
}
