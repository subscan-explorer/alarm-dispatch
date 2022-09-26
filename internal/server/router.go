package server

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func Route() http.Handler {
	srv := http.NewServeMux()
	srv.HandleFunc("/alert_manager", Webhook)
	srv.Handle("/metrics", promhttp.Handler())

	srv.HandleFunc("/healthz", func(rsp http.ResponseWriter, _ *http.Request) {
		rsp.WriteHeader(http.StatusOK)
	})
	srv.HandleFunc("/readiness", func(rsp http.ResponseWriter, _ *http.Request) {
		rsp.WriteHeader(http.StatusOK)
	})
	return srv
}
