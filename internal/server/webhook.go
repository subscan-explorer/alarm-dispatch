package server

import (
	"encoding/json"
	"net/http"

	"github.com/subscan-explorer/alarm-dispatch/internal/dispatch"
	"github.com/subscan-explorer/alarm-dispatch/internal/metrics"
	"github.com/subscan-explorer/alarm-dispatch/internal/model"
)

func Webhook(rsp http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		rsp.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var (
		alert model.Notification
		err   error
	)
	if err = json.NewDecoder(req.Body).Decode(&alert); err != nil {
		rsp.WriteHeader(http.StatusBadRequest)
		_, _ = rsp.Write([]byte(err.Error()))
		return
	}
	metrics.AddAlertStatusCount("receive", len(alert.Alerts))
	go dispatch.Dispatch(alert)
	rsp.WriteHeader(http.StatusOK)
}
