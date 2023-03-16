package internal

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func Serve() {
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":6999", nil)
}
