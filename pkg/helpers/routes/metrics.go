package routes

import (
	"net/http"

	"github.com/gorilla/mux"
)

// AddMetrics adds the common prometheus metrics routes to a given router:
//
// - ``/metrics``
//
// ```eval_rst
//
// .. note::
//    ``Subrouter`` can be used with this route.
//
//    Example:
//
//    .. code-block:: go
//
//       AddMetrics(
//         mux.NewRouter().PathPrefix("/somepath/").Subrouter(),
//         yourPrometheusExporter,
//       )
//
//    by doing so your routes will become:
//
//    - ``somepath/metrics``
// ```
func AddMetrics(router *mux.Router, prometheusExporter http.Handler) {
	// setup metrics exporter
	router.Path("/metrics").Methods("GET").Handler(prometheusExporter)
}
