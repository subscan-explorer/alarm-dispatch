package discord

import (
	"context"
	"log"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/subscan-explorer/alarm-dispatch/conf"
	"github.com/subscan-explorer/alarm-dispatch/internal/model"
)

var senders map[string]*discordgo.Session
var once sync.Once

// Notifier implements a Notifier for Slack notifications.
type Notifier struct {
	Channel       string
	cli           *discordgo.Session
	lastMessageID [2]string // 0=last message 1=need delete message
}

// New returns a new Slack notification handler.
func New(c conf.Receiver) *Notifier {
	notifier := new(Notifier)
	if len(c.ChatID) != 0 && len(c.Sender) != 0 {
		once.Do(initSender)
		notifier.Channel = c.ChatID
		notifier.cli = senders[c.Sender]
		if notifier.cli == nil {
			return nil
		}
	}
	return notifier
}

func initSender() {
	senders = make(map[string]*discordgo.Session)
	for _, s := range conf.Conf.Discord {
		c, err := discordgo.New("Bot " + s.Token)
		if err != nil {
			log.Fatalln("discord auth configure err")
		}
		senders[s.Name] = c
	}
}

func (n *Notifier) RemoveLastMessage(context.Context) {
	if len(n.lastMessageID[1]) != 0 {
		if err := n.cli.ChannelMessageDelete(n.Channel, n.lastMessageID[1]); err != nil {
			log.Printf("failed to delete slack message [%s],err: %s\n", n.lastMessageID, err.Error())
		}
	}
	n.lastMessageID[1] = ""
}

// Notify implements the Notifier interface.
func (n *Notifier) Notify(ctx context.Context, alert model.Alert) (bool, error) {
	if alert.IsResolved() {
		if len(n.lastMessageID[0]) != 0 {
			n.lastMessageID[1], n.lastMessageID[0] = n.lastMessageID[0], "nil"
		}
		return false, nil
	}
	m, err := n.cli.ChannelMessageSend(n.Channel, alert.Markdown(), discordgo.WithContext(ctx), discordgo.WithRetryOnRatelimit(true), discordgo.WithRestRetries(3))
	if err != nil {
		return true, err
	}
	if len(n.lastMessageID[0]) != 0 {
		n.lastMessageID[1] = n.lastMessageID[0]
	}
	n.lastMessageID[0] = m.ID
	return false, nil
}
