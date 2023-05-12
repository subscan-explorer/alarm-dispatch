package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"

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
	lastMessageID [2]int
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

func (n *Notifier) RemoveLastMessage(context.Context) {
	if n.lastMessageID[1] != 0 {
		if err := n.deleteMessage(n.lastMessageID[1]); err != nil {
			log.Printf("failed to delete telegram message, err: %s\n", err.Error())
		}
	}
	n.lastMessageID[1] = 0
}

func (n *Notifier) Notify(ctx context.Context, alert model.Alert) (bool, error) {
	if alert.IsResolved() {
		if n.lastMessageID[0] != 0 {
			n.lastMessageID[1], n.lastMessageID[0] = n.lastMessageID[0], 0
		}
		return false, nil
	}
	var (
		reqBody, _ = json.Marshal(n.buildMessage(alert))
		req        *http.Request
		host       = n.conf.Webhook + "sendMessage"
		rsp        *http.Response
		err        error
	)
	if req, err = http.NewRequestWithContext(ctx, http.MethodPost, host, bytes.NewReader(reqBody)); err != nil {
		return true, err
	}
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	if rsp, err = cli.HTTPCli.Do(req); err != nil {
		return true, err
	}
	defer func() {
		_, _ = io.Copy(io.Discard, rsp.Body)
		rsp.Body.Close()
	}()
	if rsp.StatusCode != 200 {
		return false, errors.New(rsp.Status)
	}
	var data SendResponse
	if err = json.NewDecoder(rsp.Body).Decode(&data); err != nil {
		return true, err
	}
	if !data.Ok {
		return false, fmt.Errorf("code: %d, message: %s", data.ErrorCode, data.Description)
	}
	if n.lastMessageID[0] != 0 {
		n.lastMessageID[1] = n.lastMessageID[0]
	}
	n.lastMessageID[0] = data.Result.MessageID
	return false, nil
}

func (n *Notifier) deleteMessage(messageID int) error {
	var host = n.conf.Webhook + "deleteMessage"
	u, _ := url.Parse(host)
	v := url.Values{}
	v.Add("chat_id", n.conf.ChatID)
	v.Add("message_id", strconv.Itoa(messageID))
	u.RawQuery = v.Encode()
	rsp, err := cli.HTTPCli.Get(u.String())
	if err != nil {
		return err
	}
	defer func() {
		_, _ = io.Copy(io.Discard, rsp.Body)
		rsp.Body.Close()
	}()
	if rsp.StatusCode > 299 {
		return errors.New(rsp.Status)
	}
	return nil
}

func (n *Notifier) buildMessage(alert model.Alert) Message {
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

type SendResponse struct {
	Ok          bool   `json:"ok,omitempty"`
	ErrorCode   int    `json:"error_code,omitempty"`
	Description string `json:"description,omitempty"`
	Result      struct {
		MessageID int `json:"message_id,omitempty"`
	} `json:"result,omitempty"`
}
