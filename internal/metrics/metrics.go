package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	sendCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "alarm",
		Subsystem: "dispatch",
		Name:      "channel_send_total",
	}, []string{"receiver", "status"})
	alertCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "alarm",
		Subsystem: "process",
		Name:      "alert_total",
	}, []string{"status"})
)

func init() {
	prometheus.MustRegister(sendCount, alertCount)
}

func IncChannelSendCount(tp, status string) {
	sendCount.WithLabelValues(tp, status).Inc()
}

func AddAlertStatusCount(status string, i int) {
	alertCount.WithLabelValues(status).Add(float64(i))
}

func IncAlertStatusCount(status string) {
	alertCount.WithLabelValues(status).Inc()
}
