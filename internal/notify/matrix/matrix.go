package matrix

import (
	"context"
	"log"
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
	lastEventID [2]string
	cli         *mautrix.Client
}

func (s *Sender) RemoveLastMessage(context.Context) {
	if len(s.lastEventID[1]) != 0 {
		if _, err := s.cli.RedactEvent(s.conf.RoomID, id.EventID(s.lastEventID[1])); err != nil {
			log.Printf("failed to delete matrix message: %s, err: %s", s.lastEventID[1], err.Error())
		}
	}
	s.lastEventID[1] = ""
}

func (s *Sender) Notify(_ context.Context, alert model.Alert) (bool, error) {
	if alert.IsResolved() {
		if len(s.lastEventID[0]) != 0 {
			s.lastEventID[1], s.lastEventID[0] = s.lastEventID[0], ""
		}
		return false, nil
	}
	rsp, err := s.cli.SendMessageEvent(s.conf.RoomID, event.EventMessage, &event.MessageEventContent{
		MsgType:       event.MsgText,
		Format:        event.FormatHTML,
		FormattedBody: alert.HTML("<br>", "&nbsp;"),
	})
	if err != nil {
		return true, err
	}
	if len(s.lastEventID[0]) != 0 {
		s.lastEventID[1] = s.lastEventID[0]
	}
	s.lastEventID[0] = rsp.EventID.String()
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
