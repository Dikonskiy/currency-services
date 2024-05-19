package metrics

import "github.com/prometheus/client_golang/prometheus"

type Metrics struct {
	DeleteDuration *prometheus.HistogramVec
	DeleteCount    *prometheus.CounterVec
}

func NewMetrics() *Metrics {
	deleteDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_delete_duration_seconds",
			Help:    "Duration of database delete queries in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "status"},
	)

	deleteCount := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "db_delete_count_total",
			Help: "Total number of database delete queries",
		},
		[]string{"method", "status"},
	)

	prometheus.MustRegister(deleteDuration)
	prometheus.MustRegister(deleteCount)

	return &Metrics{
		DeleteDuration: deleteDuration,
		DeleteCount:    deleteCount,
	}
}

func (m *Metrics) IncDeleteCount(method, status string) {
	m.DeleteCount.WithLabelValues(method, status).Inc()
}

func (m *Metrics) ObserveDeleteDuration(method, status string, duration float64) {
	m.DeleteDuration.WithLabelValues(method, status).Observe(duration)
}
