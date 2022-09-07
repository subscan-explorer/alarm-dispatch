package server

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/subscan-explorer/alarm-dispatch/internal/dispatch"
	"github.com/subscan-explorer/alarm-dispatch/internal/metrics"
	"github.com/subscan-explorer/alarm-dispatch/internal/model"
)

func Webhook(rsp http.ResponseWriter, req *http.Request) {
	//ctx := req.Context()
	if req.Method != http.MethodPost {
		rsp.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var (
		alert   model.Notification
		reqBody []byte
		err     error
	)
	if reqBody, err = io.ReadAll(req.Body); err != nil {
		rsp.WriteHeader(http.StatusBadRequest)
		_, _ = rsp.Write([]byte(err.Error()))
		return
	}
	if err = json.Unmarshal(reqBody, &alert); err != nil {
		rsp.WriteHeader(http.StatusBadRequest)
		_, _ = rsp.Write([]byte(err.Error()))
		return
	}
	log.Printf("raw data: %s\n", string(reqBody))
	//if err = json.NewDecoder(req.Body).Decode(&alert); err != nil {
	//	rsp.WriteHeader(http.StatusBadRequest)
	//	_, _ = rsp.Write([]byte(err.Error()))
	//	return
	//}
	metrics.AddAlertStatusCount("receive", len(alert.Alerts))
	go dispatch.Dispatch(alert)
	log.Printf("alert: %+v", alert)
	rsp.WriteHeader(http.StatusOK)
}
