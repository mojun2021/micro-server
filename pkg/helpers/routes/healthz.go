package routes

import (
	"net/http"

	"github.com/gorilla/mux"
)

func AddHealthz(router *mux.Router, liveness, readiness http.Handler) {
	// setup health checks, /healthz route is taken by health checks by default.
	s := router.PathPrefix("/healthz/").Subrouter()
	s.Path("/liveness").Methods("GET").Handler(liveness)
	s.Path("/readiness").Methods("GET").Handler(readiness)
}
