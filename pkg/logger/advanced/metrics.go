package logger

import (
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

const (
	// NameLogTagKeyName defines the context key name for the logger name.
	NameLogTagKeyName = "logger_name"
	// LevelLogTagKeyName defines the context key name for the logger level.
	LevelLogTagKeyName = "level"
)

var (
	// Counts the number of error logs.
	mLogs = stats.Int64("micro-server/logs", "The number of logs encountered", "1")

	nameKey, _  = tag.NewKey(NameLogTagKeyName)
	levelKey, _ = tag.NewKey(LevelLogTagKeyName)

	// LogCountView is the number of logs view. It has 2 tags, ``level`` and ``logger_name``.
	LogCountView = &view.View{
		Name:        "micro-server/logs",
		Measure:     mLogs,
		Description: "The number of logs encountered",
		Aggregation: view.Count(),
		TagKeys:     []tag.Key{nameKey, levelKey},
	}
)
