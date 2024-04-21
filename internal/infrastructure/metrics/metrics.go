package metrics

import (
	"github.com/NikitaTsaralov/transactional-outbox/config"
	"github.com/NikitaTsaralov/utils/logger"
	"github.com/prometheus/client_golang/prometheus"
)

const unprocessedEventsCounter = "unprocessed_events_counter"

type Metrics struct {
	cfg      *config.Metrics
	registry *prometheus.Registry

	//	metrics
	unprocessedEventsCounter *prometheus.GaugeVec
}

func NewMetrics(cfg *config.Metrics, registry *prometheus.Registry) *Metrics {
	metrics := &Metrics{
		cfg:      cfg,
		registry: registry,
	}

	metrics.unprocessedEventsCounter = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: prometheus.BuildFQName(metrics.cfg.Namespace, "", unprocessedEventsCounter),
	},
		[]string{},
	)

	err := metrics.registry.Register(metrics.unprocessedEventsCounter)
	if err != nil {
		logger.Fatalf("can't register metrics: %v", err)
	}

	return metrics
}

func (m *Metrics) IncUnprocessedEventsCounter(n int) {
	m.unprocessedEventsCounter.WithLabelValues().Inc()
}

func (m *Metrics) DecUnprocessedEventsCounter(n int) {
	m.unprocessedEventsCounter.WithLabelValues().Desc()
}
