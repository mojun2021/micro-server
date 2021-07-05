// Package logs contains all the advanced configuration of the different logger used in the logger module. However,
// some configurations are available for the more advanced users to modify. This is highly inspired by
// kubernetes controller-runtime logger <https://github.com/kubernetes-sigs/controller-runtime/blob/master/pkg/runtime/log/log.go>
package logger

import (
	"io"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	// LogLevelEnvironmentVariable defines the name of the environment variable to set in order to define a log level
	// different from the default one. The available log levels are:
	//
	// - debug
	// - info
	// - warning
	// - error
	//
	// This environment variable also support value level, just like the logger ``V(level int8)`` method. i.e. if your
	// logger uses ``logger.V(17).Info("my message")`` you can set ``USGO_LOG_LEVEL=17``.
	LogLevelEnvironmentVariable = "USGO_LOG_LEVEL"

	// Production logger constant
	//

	// ProductionLoggerDefaultLevel defines the default level of the production logger.
	ProductionLoggerDefaultLevel = zap.InfoLevel

	// Development logger constant
	//

	// DevelopmentLoggerDefaultLevel defines the default level of the development logger.
	DevelopmentLoggerDefaultLevel = zap.DebugLevel
)

var (
	// Global configuration
	//

	// DisableLogTime disable the time in the logs. Use this in test to avoid changing output.
	DisableLogTime = false

	// LoggerStackTraceLevel defines the default level of logger where a stack trace
	// is appended to the log.
	LoggerStackTraceLevel = zap.ErrorLevel

	// Production logger configuration
	//

	// ProductionLoggerSamplerEnabled enables the production logger log sampler. For more information on logger sampler
	// see https://godoc.org/go.uber.org/zap/zapcore#NewSampler.
	ProductionLoggerSamplerEnabled = true
	// ProductionLoggerSamplerPeriod defines the production logger sampling period. i.e.
	// The period of time the sampling parameter are based on.
	ProductionLoggerSamplerPeriod = time.Second
	// ProductionLoggerSamplerFirst defines the number of logs that are taken during the sampling period before the
	// sampler starts to sample logs.
	ProductionLoggerSamplerFirst = 100
	// ProductionLoggerSamplerThereAfter defines the number of logs that are skip once the number of log defines
	// by ProductionLoggerSamplerFirst are reached for the remaining of the sampling period.
	ProductionLoggerSamplerThereAfter = 100
)

func makeTimeEncoder(f func(t time.Time, enc zapcore.PrimitiveArrayEncoder)) func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	if DisableLogTime {
		return disabledTimeEncoder
	}
	return f
}

func disabledTimeEncoder(_ time.Time, _ zapcore.PrimitiveArrayEncoder) {}

func simpleTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("15:04:05"))
}

func productionLoggerSamplerCore(core zapcore.Core) zapcore.Core {
	return zapcore.NewSamplerWithOptions(
		core,
		ProductionLoggerSamplerPeriod,
		ProductionLoggerSamplerFirst,
		ProductionLoggerSamplerThereAfter,
	)
}

// NewProductionLogger creates a new logger to use within a cluster.
//
// A production logger is a JSON formatted logger with a default log level set to ``info``. It defines a log sampler :
//
// A logger samples by logging the first N entries (``ProductionLoggerSamplerFirst``) with a given level and message
// each tick (``ProductionLoggerSamplerPeriod``). If more Entries with the same level and message are seen during
// the same interval, every Mth message (``ProductionLoggerSamplerThereAfter``) is logged and the rest are dropped.
//
// You can deactivate the sampler with ``ProductionLoggerSamplerEnabled``
//
// Example:
//
//    {"level":"info","ts":1542211325.6108115,"logger":"sample.server","msg":"Starting the HTTP server","endpoint":":8080","url":"http://MTL-BH846:8080"}
func NewProductionLogger(destWriter io.Writer) logr.Logger {
	sink := zapcore.AddSync(destWriter)

	encCfg := zap.NewProductionEncoderConfig()

	// Production logger doesn't use the EncodeTime methods as the Development logger. To disable the
	// time within log we have to clear the TimeKey.
	if DisableLogTime {
		encCfg.TimeKey = ""
	}

	// Production logger use JSON format
	enc := zapcore.NewJSONEncoder(encCfg)

	options := []zap.Option{
		zap.AddStacktrace(LoggerStackTraceLevel),
		zap.AddCallerSkip(1),
		zap.ErrorOutput(sink),
	}

	if ProductionLoggerSamplerEnabled {
		options = append(options, zap.WrapCore(productionLoggerSamplerCore))
	}

	core := zapcore.NewCore(
		enc,
		sink,
		getLoggerLevel(ProductionLoggerDefaultLevel),
	)
	log := zap.New(core, options...)

	// return ToMonitoredLogger(zapr.NewLogger(log))
	return zapr.NewLogger(log)
}

// NewDevelopmentLogger creates a new logger to use within a developer command shell.
//
// A development logger is a console formatted logger with a default log level set to ``debug``.
//
// Example:
//
//    10:59:13        INFO    sample.server   Starting the HTTP server        {"endpoint": ":8080", "url": "http://MTL-BH846:8080"}
func NewDevelopmentLogger(destWriter io.Writer) logr.Logger {
	sink := zapcore.AddSync(destWriter)

	encCfg := zap.NewDevelopmentEncoderConfig()
	encCfg.EncodeTime = makeTimeEncoder(simpleTimeEncoder)

	// Development logger use Console encoder
	enc := zapcore.NewConsoleEncoder(encCfg)

	log := zap.New(
		zapcore.NewCore(enc, sink, getLoggerLevel(DevelopmentLoggerDefaultLevel)),
		zap.Development(),
		zap.AddStacktrace(LoggerStackTraceLevel),
		zap.AddCallerSkip(1),
		zap.ErrorOutput(sink),
	)

	//	return ToMonitoredLogger(zapr.NewLogger(log))
	return zapr.NewLogger(log)
}

func getLoggerLevel(defaultLogLevel zapcore.Level) zap.AtomicLevel {
	lvlStr := strings.ToLower(os.Getenv(LogLevelEnvironmentVariable))
	switch lvlStr {
	case "debug":
		return zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		return zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warning":
		return zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		return zap.NewAtomicLevelAt(zap.ErrorLevel)
	default:
		lvl, err := strconv.Atoi(lvlStr)
		if err == nil {
			// convert level from zap logger to logr (logger used by zap), for some reason zap V()
			// method use an inverted level handling.
			convLvl := -1 * lvl
			// max level (int8)
			if convLvl >= math.MaxInt8 {
				convLvl = math.MaxInt8
			}

			// min level (int8)
			if convLvl <= math.MinInt8 {
				convLvl = math.MinInt8
			}

			return zap.NewAtomicLevelAt(zapcore.Level(convLvl))
		}
		return zap.NewAtomicLevelAt(defaultLogLevel)
	}
}
