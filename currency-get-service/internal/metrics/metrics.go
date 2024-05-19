package metrics

import "github.com/prometheus/client_golang/prometheus"

type Metrics struct {
	SelectDuration *prometheus.HistogramVec
	SelectCount    *prometheus.CounterVec
}

func NewMetrics() *Metrics {
	selectDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_select_duration_seconds",
			Help:    "Duration of database select queries in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "status"},
	)

	selectCount := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "db_select_count_total",
			Help: "Total number of database select queries",
		},
		[]string{"method", "status"},
	)

	insertDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_insert_duration_seconds",
			Help:    "Duration of database insert queries in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "status"},
	)

	prometheus.MustRegister(selectDuration)
	prometheus.MustRegister(selectCount)
	prometheus.MustRegister(insertDuration)

	return &Metrics{
		SelectDuration: selectDuration,
		SelectCount:    selectCount,
	}
}

func (m *Metrics) ObserveSelectDuration(method, status string, duration float64) {
	m.SelectDuration.WithLabelValues(method, status).Observe(duration)
}

func (m *Metrics) IncSelectCount(method, status string) {
	m.SelectCount.WithLabelValues(method, status).Inc()
}
