package http

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

func InitMetrics(port string) {
	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		server := &http.Server{
			Addr:              ":" + port,
			Handler:           mux,
			ReadHeaderTimeout: 5 * time.Second,
			ReadTimeout:       10 * time.Second,
			WriteTimeout:      10 * time.Second,
			IdleTimeout:       60 * time.Second,
		}
		if err := server.ListenAndServe(); err != nil {
			logrus.WithError(err).Error("metrics server failed")
		}
	}()
}
