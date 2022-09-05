package server

import "net/http"

func Route() http.Handler {
	srv := http.NewServeMux()
	srv.HandleFunc("/alert_manager", Webhook)
	return srv
}
