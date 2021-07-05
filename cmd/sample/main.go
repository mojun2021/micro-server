package main

import (
	"context"
	"os"

	"golang.org/x/sync/errgroup"

	"github.com/mojun2021/micro-server/pkg/logger"
	"github.com/mojun2021/micro-server/pkg/metrics"

	ctxhelp "github.com/mojun2021/micro-server/pkg/helpers/context"
	"github.com/mojun2021/micro-server/pkg/server"
)

var appLog = logger.NewLogger(os.Stdout, "sample")

func main() {
	logger.SetLogger(appLog)

	p, err := metrics.NewPrometheusExporter()

	if err != nil {
		appLog.Error(err, "Failed")
	}

	s, err := server.NewMonitoringServer(":8080", server.Options{
		EnableProfiling: true,
	}, nil, nil, p)

	if err != nil {
		appLog.Error(err, "Failed")
	}

	//	routes.AddDebugPanel(s.Router())

	ctx, cancel := ctxhelp.WithCancelOnTermination(context.Background())
	wg, ctx := errgroup.WithContext(ctx)
	defer cancel()

	wg.Go(func() error { return s.Run(ctx) })

	// wait for all go routines to complete
	if err := wg.Wait(); err != nil {
		appLog.Error(err, "Unhandled error received")
	}
}
