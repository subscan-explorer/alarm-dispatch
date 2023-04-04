package dispatch

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/subscan-explorer/alarm-dispatch/internal/metrics"
	"github.com/subscan-explorer/alarm-dispatch/internal/model"
	"github.com/subscan-explorer/alarm-dispatch/internal/notify"
)

type Processor interface {
	Filter(...model.Alert) []model.Alert
	Replace(...model.Alert) []model.Alert
	Dispatch(...model.Alert) []model.Alert
}

func NewProcess() Processor {
	return InitProcess()
}

func Dispatch(alert model.Notification) {
	pr := NewProcess()
	alert.Alerts = pr.Filter(alert.Alerts...)
	alert.Alerts = pr.Replace(alert.Alerts...)
	alert.Alerts = pr.Dispatch(alert.Alerts...)
	// group
	var wg = new(sync.WaitGroup)

	for _, a := range alert.Alerts {
		for _, r := range a.Receiver {
			if n := notify.GetNoticer(r); n != nil {
				wg.Add(1)
				go func(name string, alert model.Alert) {
					for i := 0; i < 3; i++ {
						metrics.IncChannelSendCount(name, "send")
						retry, err := n.Notify(context.Background(), alert)
						if err != nil {
							metrics.IncChannelSendCount(name, "failed")
							log.Println("err: ", err.Error())
						} else {
							metrics.IncChannelSendCount(name, "success")
						}
						if retry {
							time.Sleep(time.Second)
							continue
						}
						// Delete after sending successfully
						n.RemoveLastMessage(context.Background())
						break
					}
					wg.Done()
				}(r, a)
			}
		}
	}
	wg.Wait()
	log.Println("complete send")
}
