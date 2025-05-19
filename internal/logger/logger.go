package logger

import (
	"go.uber.org/zap"
)

// Log — глобальный логгер, используемый в приложении.
// По умолчанию инициализирован как noop (не выполняет логирование).
var Log *zap.Logger = zap.NewNop()

// Initialize инициализирует глобальный логгер с заданным уровнем логирования.
// Принимает строку level — уровень логирования (например, "debug", "info", "warn", "error").
// Возвращает ошибку, если уровень не может быть распознан.
// После вызова Log будет настроен на указанный уровень и готов к использованию.
func Initialize(level string) error {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}
	cfg := zap.NewProductionConfig()
	cfg.Level = lvl
	Log, _ = cfg.Build()
	return nil
}
