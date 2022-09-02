package dispatch

import (
	"context"
	"log"

	"github.com/subscan-explorer/alarm-dispatch/internal/model"
	"github.com/subscan-explorer/alarm-dispatch/internal/notify"
)

type Processor interface {
	Filter(...model.Alert) []model.Alert
	Dispatch(...model.Alert) []model.Alert
	// LabelReplace
	// LabelFilter()
	// AnnotationsFilter()
}

func NewProcess() Processor {
	return InitProcess()
}

func Dispatch(alert model.Notification) {
	pr := NewProcess()
	alert.Alerts = pr.Filter(alert.Alerts...)
	alert.Alerts = pr.Dispatch(alert.Alerts...)
	log.Printf("alert: %+v\n", alert)
	// group
	for _, a := range alert.Alerts {
		for _, r := range a.Receiver {
			if n := notify.GetNoticer(r); n != nil {
				_, err := n.Notify(context.Background(), map[string]string{
					"summary": alert.CommonAnnotations["summary"],
				}, a)
				if err != nil {
					log.Println("err: ", err.Error())
				}
			}
		}
	}

}
