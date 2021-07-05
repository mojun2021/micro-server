// Package handlers contains a list of default handlers that can be used with the different routes.
//
// Examples:
//
// ```go
//
// advserver := advserver.NewBaseServer(":8080", time.Minute)
//
// advserver.Router().Path("/somepath").Method("GET").Handler(StatusOkHandler)
//
// ```
package handlers

import (
	"net/http"

	"gocloud.dev/server/health"
)

var (
	// StatusOkHandler is a simple HandlerFunc that always respond ``StatusOk``. It also prompt ``ok`` into your browser.
	StatusOkHandler = http.HandlerFunc(health.HandleLive)
)
