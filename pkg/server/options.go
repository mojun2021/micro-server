package server

import (
	"time"
)

var (
	defaultGracefulTimeout = time.Second * 5
)

// Options represents the server configuration options.
type Options struct {
	// When `true`, enables profiling support and exposes a `debug` endpoint.
	EnableProfiling bool
	// When `true`, enables tracing support on the server requests.
	EnableTracing bool
	// GracefulTimeout is the timeout duration for the server graceful shutdown.
	GracefulTimeout time.Duration
	// EnableReplication enables header replication support on the server responses.
	EnableReplication bool
}

func setOptionsDefaults(options *Options) {
	if options != nil {
		if options.GracefulTimeout == 0 {
			options.GracefulTimeout = defaultGracefulTimeout
		}
	}
}
