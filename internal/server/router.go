package server

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func Route() http.Handler {
	srv := http.NewServeMux()
	srv.HandleFunc("/alert_manager", Webhook)
	srv.Handle("/metrics", promhttp.Handler())
	return srv
}
