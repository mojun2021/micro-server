package trace

import (
	"net/url"
	"os"
	"strings"

	ocjaeger "contrib.go.opencensus.io/exporter/jaeger"

	"github.com/mojun2021/micro-server/pkg/logger"
)

var log = logger.Log.WithName("trace")

const (
	jaegerAgentHostEnvKey     = "JAEGER_AGENT_HOST"
	jaegerAgentPort           = "6831"
	jaegerCollectorHostEnvKey = "JAEGER_COLLECTOR_HOST"
	jaegerCollectorPath       = "/api/traces"
	jaegerCollectorPort       = "14268"
	jaegerServiceNameEnvKey   = "JAEGER_SERVICE_NAME"
)

// RegisterJaegerExporter registers the tracing exporter.
//
// Some configuration can be made through Environment variables:
//
// - `JAEGER_AGENT_HOST` : sets the jaeger agent host
//
// - `JAEGER_COLLECTOR_HOST` : sets the jaeger collector host
//
// - `JAEGER_SERVICE_NAME` : sets the jaeger service name
func RegisterJaegerExporter(r JaegerRegisterOptions) error {
	setJaegerRegisterOptionsDefault(&r)

	exporter, options, err := r.NewJaegerExporter()
	if err != nil {
		log.Info("Failed to retrieve Jaeger tracing exporter", "error", err)
		return err
	}

	r.RegisterExporter(exporter)

	log.Info(
		"Registered Jaeger exporter",
		"agent_endpoint", options.AgentEndpoint,
		"collector_endpoint", options.CollectorEndpoint,
	)

	return nil
}

func newJaegerExporter() (*ocjaeger.Exporter, *ocjaeger.Options, error) {
	jaegerServiceName := os.Getenv(jaegerServiceNameEnvKey)
	if jaegerServiceName == "" {
		log.Info("JAEGER_SERVICE_NAME not defined, using default service name")
	}

	options := ocjaeger.Options{
		Process: ocjaeger.Process{
			ServiceName: jaegerServiceName,
		},
	}
	setJaegerOptionsEndpoint(&options)

	exporter, err := ocjaeger.NewExporter(options)

	return exporter, &options, err
}

func setJaegerOptionsEndpoint(options *ocjaeger.Options) {
	// Set Jaeger Collector endpoint
	if jaegerCollectorHost := os.Getenv(jaegerCollectorHostEnvKey); jaegerCollectorHost != "" {
		endpointURL := &url.URL{
			Scheme: "http",
			Host:   strings.Join([]string{jaegerCollectorHost, jaegerCollectorPort}, ":"),
			Path:   jaegerCollectorPath,
		}
		options.CollectorEndpoint = endpointURL.String()

		return
	}

	log.Info("JAEGER_COLLECTOR_HOST not defined, using the agent endpoint instead.")

	// Set Jaeger Agent endpoint
	hostname := "localhost"
	if jaegerAgentHost := os.Getenv(jaegerAgentHostEnvKey); jaegerAgentHost != "" {
		hostname = jaegerAgentHost

	} else {
		log.Info("JAEGER_AGENT_HOST not defined, using default agent host")
	}

	options.AgentEndpoint = strings.Join([]string{hostname, jaegerAgentPort}, ":")
}
