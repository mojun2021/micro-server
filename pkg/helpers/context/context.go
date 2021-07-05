package context

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// WithCancelOnSignal returns a context that will get cancelled whenever one of
// the specified signals is caught.
func WithCancelOnSignal(ctx context.Context, signals ...os.Signal) (context.Context, func()) {
	var once sync.Once
	ctx, cancel := context.WithCancel(ctx)

	ch := make(chan os.Signal, len(signals))

	signal.Notify(ch, signals...)

	go func() {
		<-ch
		cancel()
	}()

	return ctx, func() {
		once.Do(func() {
			signal.Reset(signals...)
			close(ch)
		})
		<-ctx.Done()
	}
}

// WithCancelOnTermination returns a context that will get cancelled whenever
// an interruption signal is caught.
func WithCancelOnTermination(ctx context.Context) (context.Context, func()) {
	return WithCancelOnSignal(ctx, os.Interrupt, syscall.SIGTERM)
}
