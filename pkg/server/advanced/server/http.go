package server

import (
	"context"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/Microsoft/go-winio"
	"github.com/gorilla/handlers"
)

// CORSOptions contains all the CORS option used by the exposed APIs.
var CORSOptions = []handlers.CORSOption{
	handlers.AllowedOrigins([]string{"*"}),
	handlers.AllowedMethods([]string{"GET", "PUT", "POST", "DELETE", "OPTIONS", "HEAD", "PATCH"}),
	handlers.AllowCredentials(),
}

// NewServer instantiates a new HTTP server with appropriate defaults.
func NewServer(handler http.Handler, listener net.Listener) *http.Server {
	return &http.Server{
		Addr:         listener.Addr().String(),
		Handler:      ConfigureHandler(handler),
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 10,
	}
}

// NewListener instantiates a new net listener.
func NewListener(endpoint string) (net.Listener, error) {
	return parseListener(endpoint)
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

// ConfigureHandler the specified handler with the default CORS, HTTP method
// override and debug routes support.
func ConfigureHandler(handler http.Handler) http.Handler {
	handler = handlers.CORS(CORSOptions...)(handler)
	handler = handlers.HTTPMethodOverrideHandler(handler)

	return handler
}

// RunServer runs a http server for as long as the specified context remains
// valid.
//
// It waits as long as gracefulTimeout for the HTTP server to exit gracefully.
func RunServer(ctx context.Context, httpServer *http.Server, listener net.Listener, gracefulTimeout time.Duration) (err error) {
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), gracefulTimeout)
		defer cancel()

		_ = httpServer.Shutdown(shutdownCtx)
		_ = httpServer.Close()
	}()

	err = httpServer.Serve(listener)

	// This happens during a graceful shutdown and is not an error.
	if err == http.ErrServerClosed {
		err = nil
	}

	return
}
