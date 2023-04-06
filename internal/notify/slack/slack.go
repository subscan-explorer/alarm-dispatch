package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/slack-go/slack"
	"github.com/subscan-explorer/alarm-dispatch/conf"
	"github.com/subscan-explorer/alarm-dispatch/internal/model"
	"github.com/subscan-explorer/alarm-dispatch/internal/notify/cli"
)

var senders map[string]*slack.Client
var once sync.Once

// Notifier implements a Notifier for Slack notifications.
type Notifier struct {
	Webhook       string
	Channel       string
	cli           *slack.Client
	lastMessageID [2][]string // 0=last message 1=need delete message
}

// New returns a new Slack notification handler.
func New(c conf.Receiver) *Notifier {
	notifier := new(Notifier)
	if len(c.Webhook) != 0 {
		notifier.Webhook = c.Webhook
		return notifier
	}
	if len(c.ChatID) != 0 && len(c.Sender) != 0 {
		once.Do(initSender)
		notifier.Channel = c.ChatID
		notifier.cli = senders[c.Sender]
	}
	return notifier
}

func initSender() {
	senders = make(map[string]*slack.Client)
	for _, s := range conf.Conf.Slack {
		c := slack.New(s.Token)
		if _, err := c.AuthTest(); err != nil {
			log.Fatalln("slack auth configure err")
		}
		senders[s.Name] = c
	}
}

func (n *Notifier) RemoveLastMessage(context.Context) {
	if len(n.lastMessageID[1]) == 2 {
		if _, _, err := n.cli.DeleteMessage(n.lastMessageID[1][0], n.lastMessageID[1][1]); err != nil {
			log.Printf("failed to delete slack message [%s],err: %s\n", n.lastMessageID, err.Error())
		}
	}
	n.lastMessageID[1] = nil
}

// Notify implements the Notifier interface.
func (n *Notifier) Notify(ctx context.Context, alert model.Alert) (bool, error) {
	if len(n.Webhook) != 0 {
		return n.SendWebhook(ctx, alert)
	}
	if n.cli != nil {
		return n.SendChannel(ctx, alert)
	}
	return false, nil
}

func (n *Notifier) SendWebhook(ctx context.Context, alert model.Alert) (bool, error) {
	var (
		reqBody, _ = json.Marshal(n.buildMessage(alert))
		req        *http.Request
		rsp        *http.Response
		err        error
	)
	if req, err = http.NewRequestWithContext(ctx, http.MethodPost, n.Webhook, bytes.NewReader(reqBody)); err != nil {
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

func (n *Notifier) SendChannel(_ context.Context, alert model.Alert) (bool, error) {
	if alert.IsResolved() {
		if len(n.lastMessageID[0]) != 0 {
			n.lastMessageID[1], n.lastMessageID[0] = n.lastMessageID[0], nil
		}
		return false, nil
	}
	msg := n.buildMessage(alert)
	channel, messageID, _, err := n.cli.SendMessage(n.Channel, slack.MsgOptionBlocks(msg.Blocks...))
	if err != nil {
		return true, err
	}
	if len(n.lastMessageID[0]) != 0 {
		n.lastMessageID[1] = n.lastMessageID[0]
	}
	n.lastMessageID[0] = []string{channel, messageID}
	return false, nil
}

func (n *Notifier) buildMessage(alert model.Alert) Message {
	var msg Message
	var blocks []*slack.TextBlockObject
	msg.Blocks = append(msg.Blocks, slack.NewHeaderBlock(slack.NewTextBlockObject("plain_text", alert.Status, false, false)))
	blocks = append(blocks, newSlackMkdBlock(fmt.Sprintf("*Start:* \n%s", alert.StartsAt.Format(time.RFC3339))))
	if !alert.EndsAt.IsZero() {
		blocks = append(blocks, newSlackMkdBlock(
			fmt.Sprintf("*End:* \n%s", alert.EndsAt.Format(time.RFC3339))))
	}
	msg.Blocks = append(msg.Blocks, slack.NewSectionBlock(nil, blocks, nil))

	for k, v := range alert.Annotations {
		if len(k) == 0 || len(v) == 0 {
			continue
		}
		msg.Blocks = append(msg.Blocks, slack.NewSectionBlock(newSlackMkdBlock(fmt.Sprintf("*%s:* \n%s", strings.Title(k), v)), nil, nil))
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
		msg.Blocks = append(msg.Blocks, slack.NewSectionBlock(newSlackMkdBlock(fmt.Sprintf("*Tag:* \n%s", buf.String())), nil, nil))
		model.PutByteBuf(buf)
	}
	return msg
}

func newSlackMkdBlock(str string) *slack.TextBlockObject {
	return slack.NewTextBlockObject("mrkdwn", str, false, false)
}

type Message struct {
	Blocks []slack.Block `json:"blocks,omitempty"`
}
