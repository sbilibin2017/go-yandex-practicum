// Package logger provides a centralized logger based on Uber's zap logging library.
package logger

import (
	"go.uber.org/zap"
)

// Log is a global sugared logger instance that can be used throughout the application.
//
// It is initialized as a no-op logger by default. Use Initialize to configure it properly.
var Log *zap.SugaredLogger = zap.NewNop().Sugar()

// Initialize sets up the global Log instance with the given logging level.
//
// The `level` parameter should be a string compatible with zap's logging levels,
// such as "debug", "info", "warn", or "error".
//
// Example:
//
//	err := logger.Initialize("debug")
//	if err != nil {
//	    panic(err)
//	}
//
// Returns an error if the level cannot be parsed.
func Initialize(level string) error {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}
	cfg := zap.NewProductionConfig()
	cfg.Level = lvl
	baseLogger, _ := cfg.Build()
	Log = baseLogger.Sugar()
	return nil
}
