package observability

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

var (
	RequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "bot",
			Name:      "handler_duration_seconds",
			Help:      "Duration of bot handlers",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"handler"},
	)
	DBErrors = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "bot",
			Name:      "db_errors_total",
			Help:      "Total number of database errors",
		},
	)
	PaymentAttempts = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "bot",
			Name:      "payment_attempts_total",
			Help:      "Number of payment attempts",
		},
	)
)

func init() {
	prometheus.MustRegister(RequestDuration, DBErrors, PaymentAttempts)
}

// Handler returns http.Handler to expose metrics.
func Handler() http.Handler { return promhttp.Handler() }
