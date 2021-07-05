// Package logs defines the default micro-server-go logger. It is highly inspired by
// [kubernetes controller-runtime logger](https://github.com/kubernetes-sigs/controller-runtime/blob/master/pkg/runtime/log/log.go)
//
//
// To use this package,
//
//     import "github.com/mojun2021/micro-server/pkg/logs"
//
//     // First create a logger
//     var appLog = logs.NewLogger(os.Stdout, "your-application-name")
//
//     // Then set the micro-server-library logger. If you do not do
//     // this, you will not receive any logs from the library. However
//     // you will still be able to use the logger.
//     logs.SetLogger(appLog)
//
//
// At any time you can change the default logging level by setting the ``USGO_LOG_LEVEL`` environment
// variable. The available log levels are:
//
// - debug
//
// - info
//
// - warning
//
// - error
//
// This environment variable also support value level, just like the logger ``V(level int8)`` method. i.e. if your
// logger uses ``logger.V(17).Info("my message")`` you can set ``USGO_LOG_LEVEL=17``.
package logger

import (
	"io"

	"github.com/go-logr/logr"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/mojun2021/micro-server/pkg/helpers/production"
	logs "github.com/mojun2021/micro-server/pkg/logger/advanced"
)

var Log = logf.Log

func NewLogger(w io.Writer, appName string) (log logr.Logger) {
	if !production.InProduction() {
		log = logs.NewDevelopmentLogger(w)
	} else {
		log = logs.NewProductionLogger(w)
	}
	return log.WithName(appName)
}

// SetLogger sets a concrete logging implementation for all deferred Loggers.
func SetLogger(l logr.Logger) { logf.SetLogger(l) }
