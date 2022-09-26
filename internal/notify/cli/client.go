package cli

import (
	"net"
	"net/http"
	"time"
)

var HTTPCli *http.Client

func init() {
	HTTPCli = &http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   time.Second * 10,
				KeepAlive: time.Second * 30,
			}).DialContext,
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     time.Second * 90,
			TLSHandshakeTimeout: time.Second * 10,
			ForceAttemptHTTP2:   true,
		},
	}
}
