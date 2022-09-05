package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/subscan-explorer/alarm-dispatch/conf"
	"github.com/subscan-explorer/alarm-dispatch/internal/dispatch"
	"github.com/subscan-explorer/alarm-dispatch/internal/notify"
	"github.com/subscan-explorer/alarm-dispatch/internal/server"
)

func main() {
	addr := flag.String("addr", ":8080", "http server listen address")
	flag.Parse()
	ctx, cancel := context.WithCancel(context.Background())
	// init config
	conf.InitConf(ctx)
	notify.Init()
	dispatch.NewProcess()

	handler := server.Route()
	srv := http.Server{Addr: *addr, Handler: handler}
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				panic(err)
			}
		}
	}()
	listenExit(cancel)
	_ = srv.Close()
	cancel()
}
func listenExit(cancel context.CancelFunc) {
	sign := make(chan os.Signal, 1)
	signal.Notify(sign, os.Kill, os.Interrupt, syscall.SIGTERM)
	s := <-sign
	log.Printf("receive signal %s, exit...\n", s.String())
	cancel()
}
