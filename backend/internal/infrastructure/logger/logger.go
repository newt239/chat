package logger

import (
	"go.uber.org/zap"
)

var globalLogger *zap.Logger

func Init(env string) error {
	var cfg zap.Config
	if env == "production" {
		cfg = zap.NewProductionConfig()
	} else {
		cfg = zap.NewDevelopmentConfig()
	}

	logger, err := cfg.Build()
	if err != nil {
		return err
	}

	globalLogger = logger
	return nil
}

func Get() *zap.Logger {
	if globalLogger == nil {
		// Fallback to a basic logger if not initialized
		globalLogger, _ = zap.NewDevelopment()
	}
	return globalLogger
}

func Sync() {
	if globalLogger != nil {
		_ = globalLogger.Sync()
	}
}
