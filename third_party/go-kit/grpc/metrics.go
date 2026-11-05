package grpc

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type serviceMetrics struct {
	activeStreams prometheus.Gauge
}

func newServiceMetrics(serviceName string) serviceMetrics {
	labels := prometheus.Labels{
		"grpc_service": serviceName,
	}

	return serviceMetrics{
		activeStreams: promauto.NewGauge(prometheus.GaugeOpts{
			Name:        "grpc_server_active_streams",
			Help:        "The total number of streams connected to the service",
			ConstLabels: labels,
		}),
	}
}
