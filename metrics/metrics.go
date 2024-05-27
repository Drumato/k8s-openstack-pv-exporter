package metrics

import "github.com/prometheus/client_golang/prometheus"

var ()

func InitializeMetrics() {
	prometheus.NewGaugeVec(prometheus.GaugeOpts{}, []string{})
}
