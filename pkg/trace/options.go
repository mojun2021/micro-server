package trace

import (
	ocjaeger "contrib.go.opencensus.io/exporter/jaeger"
	octrace "go.opencensus.io/trace"
)

// NewJaegerExporterFn defines a function that creates a new Jaeger trace exporter.
type NewJaegerExporterFn func() (*ocjaeger.Exporter, *ocjaeger.Options, error)

// RegisterExporterFn defines a function that registers a new trace exporter.
type RegisterExporterFn func(e octrace.Exporter)

// JaegerRegisterOptions are the arguments to register a Jaeger trace exporter.
type JaegerRegisterOptions struct {
	// NewJaegerExporter will create a new Jaeger trace exporter.
	NewJaegerExporter NewJaegerExporterFn
	// RegisterExporter will register a new trace exporter.
	RegisterExporter RegisterExporterFn
}

func setJaegerRegisterOptionsDefault(options *JaegerRegisterOptions) {
	if options.NewJaegerExporter == nil {
		options.NewJaegerExporter = newJaegerExporter
	}

	if options.RegisterExporter == nil {
		options.RegisterExporter = octrace.RegisterExporter
	}
}
