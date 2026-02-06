package http

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func InitMetrics(port string) {
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":"+port, nil)
	}()
}
