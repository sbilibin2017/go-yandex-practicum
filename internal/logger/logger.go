package logger

import (
	"go.uber.org/zap"
)

var Log *zap.SugaredLogger

func Initialize(level string) error {
	if Log != nil {
		return nil
	}
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}
	cfg := zap.NewProductionConfig()
	cfg.Level = lvl
	zl, _ := cfg.Build()
	Log = zl.Sugar()
	return nil
}
