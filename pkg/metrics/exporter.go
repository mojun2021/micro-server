package metrics

import (
	ocprometheus "contrib.go.opencensus.io/exporter/prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	ocview "go.opencensus.io/stats/view"

	"github.com/mojun2021/micro-server/pkg/logger"
	advlogs "github.com/mojun2021/micro-server/pkg/logger/advanced"
)

var log = logger.Log.WithName("metrics")

// NewPrometheusExporter creates a new Prometheus metrics exporter.
func NewPrometheusExporter(views ...*ocview.View) (*ocprometheus.Exporter, error) {
	// Create the Prometheus metrics registry and register collectors
	metricsRegistry, ok := prometheus.DefaultRegisterer.(*prometheus.Registry)
	if !ok {
		metricsRegistry = prometheus.NewRegistry()
		metricsRegistry.MustRegister(collectors.NewGoCollector())
	}

	return NewPrometheusExporterFromRegistry(metricsRegistry, views...)
}

// NewPrometheusExporterFromRegistry creates a new Prometheus metrics exporter.
func NewPrometheusExporterFromRegistry(registry *prometheus.Registry, views ...*ocview.View) (*ocprometheus.Exporter, error) {
	prometheusExporter, err := ocprometheus.NewExporter(
		ocprometheus.Options{
			Registry: registry,
		},
	)
	if err != nil {
		return nil, err
	}

	// register the stats exporter
	ocview.RegisterExporter(prometheusExporter)

	views = append(
		views,
		advlogs.LogCountView,
		//middlewares.ServerRequestBytesView,
		//middlewares.ServerResponseCountView,
		//middlewares.ServerResponseBytesView,
		//middlewares.ServerLatencyView,
	)

	// register the views
	if err := ocview.Register(views...); err != nil {
		log.Error(err, "Failed to register views")
	}
	return prometheusExporter, nil
}
