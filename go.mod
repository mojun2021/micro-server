module github.com/mojun2021/micro-server

go 1.15

require (
	contrib.go.opencensus.io/exporter/jaeger v0.2.1
	contrib.go.opencensus.io/exporter/prometheus v0.3.0
	github.com/Microsoft/go-winio v0.5.0
	github.com/go-logr/logr v0.4.0
	github.com/go-logr/zapr v0.4.0
	github.com/gorilla/handlers v1.5.1
	github.com/gorilla/mux v1.8.0
	github.com/prometheus/client_golang v1.11.0
	go.opencensus.io v0.23.0
	go.uber.org/zap v1.18.1
	gocloud.dev v0.23.0
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	sigs.k8s.io/controller-runtime v0.9.2
)
