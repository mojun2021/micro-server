package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	ocprometheus "contrib.go.opencensus.io/exporter/prometheus"
	"github.com/Microsoft/go-winio"
	"github.com/go-logr/logr"
	"github.com/gorilla/mux"

	"github.com/mojun2021/micro-server/pkg/helpers/handlers"
	"github.com/mojun2021/micro-server/pkg/helpers/production"
	"github.com/mojun2021/micro-server/pkg/helpers/routes"
	"github.com/mojun2021/micro-server/pkg/logger"
	advserver "github.com/mojun2021/micro-server/pkg/server/advanced/server"
)

// Server is the HTTP server interface.
type Server interface {
	// Run launches the http server.
	Run(ctx context.Context) error
	// Shutdown closes all server used resources.
	Shutdown() error
	// Router gives you the server's router. This allows you to add your specific route.
	Router() *mux.Router
	// Listener gives you the server's listener.
	Listener() net.Listener
	// GetServerURL returns the server url.
	GetServerURL() *url.URL
	// IsTracingEnabled returns `true` when the tracing support is enabled.
	IsTracingEnabled() bool
	// IsReplicatingEnabled returns `true` when the support for headers replication is enabled.
	IsReplicatingEnabled() bool
}

type httpServer struct {
	endpoint        string
	gracefulTimeout time.Duration
	router          *mux.Router
	listener        net.Listener
	serverURL       *url.URL
	logger          logr.Logger
	runningServer   *http.Server
	// telemetryOptions  *middlewares.TelemetryOptions
	// headerReplication *middlewares.ServerOptions
}

// NewBaseServer returns a new basic HTTP server without any predefined routes.
func NewBaseServer(endpoint string, options Options) (Server, error) {
	setOptionsDefaults(&options)

	listener, err := advserver.NewListener(endpoint)
	if err != nil {
		return nil, err
	}

	serverURL := url.URL{
		Scheme: "http",
		Host:   production.EndpointToHostname(listener.Addr().String(), production.InProduction()),
	}

	newLog := logger.Log.WithName("server").WithValues(
		"endpoint", endpoint,
		"url", serverURL.String(),
	)

	//var enableTracing bool
	//if options.EnableTracing {
	//	if err := trace.RegisterJaegerExporter(trace.JaegerRegisterOptions{}); err == nil {
	//		enableTracing = true
	//
	//	} else {
	//		newLog.Info("Failed to set Jaeger exporter", "error", err)
	//	}
	//}

	s := &httpServer{
		endpoint:        endpoint,
		gracefulTimeout: options.GracefulTimeout,
		router:          mux.NewRouter(),
		listener:        listener,
		serverURL:       &serverURL,
		logger:          newLog,
		runningServer:   nil,
		//telemetryOptions:  middlewares.NewTelemetryOptions(enableTracing),
		//headerReplication: middlewares.NewServerOptions(options.EnableReplication),
	}
	return s, nil
}

// Parse a listener.
//
// If the endpoint starts with "\\", a Windows named-pipe name is assumed.
//
// Otherwise, falls back to a TCP listener.
//
// An example of valid Windows named-pipe name is: \\.\pipe\MyPipe
func parseListener(endpoint string) (net.Listener, error) {
	if strings.HasPrefix(endpoint, "\\\\") {
		return winio.ListenPipe(endpoint, nil)
	}

	return net.Listen("tcp", endpoint)
}

// NewMonitoringServer returns a new HTTP server with basic monitoring routes.
//
// - ``/healthz/liveness``
//
// - ``/healthz/readiness``
//
// - ``/metrics``
//
// Simplest Example:
//
//     NewMonitoringServer(":8080", Options{}, nil, nil, nil)
//
func NewMonitoringServer(
	endpoint string,
	options Options,
	liveness http.Handler,
	readiness http.Handler,
	prometheusExporter *ocprometheus.Exporter,
) (Server, error) {
	server, err := NewBaseServer(endpoint, options)
	if err != nil {
		return nil, err
	}

	if liveness == nil {
		liveness = handlers.StatusOkHandler
	}

	if readiness == nil {
		readiness = handlers.StatusOkHandler
	}

	routes.AddHealthz(server.Router(), liveness, readiness)

	if prometheusExporter != nil {
		routes.AddMetrics(server.Router(), prometheusExporter)
	}

	return server, nil
}

// Router gives you the server's router. This allows you to add your specific route.
//
// example:
// ```go
// server.Router().Path("/route").Methods("GET").HandlerFunc(someHandlers)
// ```
func (s *httpServer) Router() *mux.Router { return s.router }

// Listener gives you access to the server net listener.
func (s *httpServer) Listener() net.Listener { return s.listener }

// Run launches the http server.
func (s *httpServer) Run(ctx context.Context) error {
	if s.listener == nil {
		return fmt.Errorf("listener not initialised")
	}

	if s.runningServer != nil {
		return fmt.Errorf("server already running")
	}

	// setup http server
	s.runningServer = advserver.NewServer(s.router, s.listener)
	defer func() { _ = s.Shutdown() }()

	s.logger.Info("Starting the HTTP server")
	defer s.logger.Info("Stopped the HTTP server")

	//if s.telemetryOptions.IsTracingEnabled() {
	//	s.logger.Info("Trace support is enabled")
	//
	//} else {
	//	s.logger.Info("Trace support is disabled")
	//}
	//
	//if s.headerReplication.IsHeaderReplicationEnabled() {
	//	s.logger.Info("Header replication support is enabled")
	//
	//} else {
	//	s.logger.Info("Header replication support is disabled")
	//}

	// Walk through all routes to log them.
	_ = s.router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		t, err := route.GetPathTemplate()

		if err != nil {
			return err
		}

		m, err := route.GetMethods()

		logger := s.logger

		if err == nil {
			logger = logger.WithValues("method", m)
		}

		host := s.GetServerURL()

		logger.Info(fmt.Sprintf("Exposed Route: `%s%s`", host, t))

		// If we found an handler, we hijack it with our metrics middleware.
		if route.GetHandler() != nil {
			handler := route.GetHandler()

			// Append middlewares to handler
			//handler = middlewares.TelemetryHandler(handler, t, s.telemetryOptions)
			//handler = middlewares.HeaderReplicatorHandler(handler, s.headerReplication)

			route.Handler(handler)
		}

		return nil
	})

	if err := advserver.RunServer(ctx, s.runningServer, s.listener, s.gracefulTimeout); err != nil {
		s.logger.Error(err, "Failed to run the HTTP server")
		return err
	}

	return nil
}

// Shutdown closes all server used resources.
func (s *httpServer) Shutdown() error {
	// If the server is currently running ... shut it down
	if s.runningServer != nil {
		err := s.runningServer.Shutdown(context.Background())

		if err != nil {
			s.logger.Error(err, "error during server shutdown")
			return err
		}

		err = s.runningServer.Close()

		if err != nil {
			s.logger.Error(err, "error during server close")
			return err
		}

		// server shutdown closes the listener
		s.listener = nil
		s.runningServer = nil
	}

	if s.listener != nil {
		err := s.listener.Close()

		if err != nil {
			s.logger.Error(err, "error during listener close")
			return err
		}

		s.listener = nil
	}
	return nil
}

// GetServerURL returns the URL on which is run the server.
func (s *httpServer) GetServerURL() *url.URL { return s.serverURL }

// IsTracingEnabled returns `true` when the tracing support is enabled.
func (s *httpServer) IsTracingEnabled() bool {
	return false
	//return s.telemetryOptions.IsTracingEnabled()
}

// IsReplicatingEnabled returns `true` when the support for headers replication is enabled.
func (s *httpServer) IsReplicatingEnabled() bool {
	return false
	//return s.headerReplication.IsHeaderReplicationEnabled()
}
