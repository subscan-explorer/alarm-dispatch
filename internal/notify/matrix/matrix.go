package matrix

import (
	"context"
	"log"
	"strings"
	"sync"

	"github.com/subscan-explorer/alarm-dispatch/conf"
	"github.com/subscan-explorer/alarm-dispatch/internal/model"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

var senders map[string]*mautrix.Client
var once sync.Once

type Sender struct {
	conf struct {
		RoomID id.RoomID
	}
	cli *mautrix.Client
}

func (s *Sender) Notify(_ context.Context, _ map[string]string, alert model.Alert) (bool, error) {
	_, err := s.cli.SendMessageEvent(s.conf.RoomID, event.EventMessage, &event.MessageEventContent{
		MsgType:       event.MsgText,
		Format:        event.FormatHTML,
		FormattedBody: strings.ReplaceAll(alert.HTML(), "\n", "<br/>"),
	})
	return false, err
}

func New(c conf.Receiver) *Sender {
	once.Do(initSender)
	s := new(Sender)
	s.conf.RoomID = id.RoomID(c.RoomID)
	if s.cli = senders[c.Sender]; s.cli == nil {
		log.Fatalln("not configure matrix sender account: ", c.Sender)
	}
	return s
}

func initSender() {
	senders = make(map[string]*mautrix.Client)
	for _, matrix := range conf.Conf.Matrix {
		var (
			cli *mautrix.Client
			err error
		)
		if cli, err = mautrix.NewClient(matrix.Host, "", ""); err != nil {
			log.Fatalln(err.Error())
		}
		_, err = cli.Login(&mautrix.ReqLogin{
			Type: "m.login.password",
			Identifier: mautrix.UserIdentifier{
				Type: mautrix.IdentifierTypeUser,
				User: matrix.User,
			},
			Password:         matrix.Password,
			StoreCredentials: true,
		})
		if err != nil {
			log.Fatalln(err.Error())
		}
		senders[matrix.Name] = cli
	}
}
