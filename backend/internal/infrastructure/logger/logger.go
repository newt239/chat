package logger

import (
	"github.com/newt239/chat/internal/domain/service"
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

// zapLogger はドメインサービスのLoggerインターフェースを実装します
type zapLogger struct {
	logger *zap.Logger
}

// NewLogger は新しいLoggerを作成します
func NewLogger() service.Logger {
	return &zapLogger{
		logger: Get(),
	}
}

// Info は情報レベルのログを出力します
func (l *zapLogger) Info(msg string, fields ...service.LogField) {
	l.logger.Info(msg, convertFields(fields)...)
}

// Warn は警告レベルのログを出力します
func (l *zapLogger) Warn(msg string, fields ...service.LogField) {
	l.logger.Warn(msg, convertFields(fields)...)
}

// Error はエラーレベルのログを出力します
func (l *zapLogger) Error(msg string, fields ...service.LogField) {
	l.logger.Error(msg, convertFields(fields)...)
}

// Debug はデバッグレベルのログを出力します
func (l *zapLogger) Debug(msg string, fields ...service.LogField) {
	l.logger.Debug(msg, convertFields(fields)...)
}

// convertFields はドメインサービスのLogFieldをzapのFieldに変換します
func convertFields(fields []service.LogField) []zap.Field {
	zapFields := make([]zap.Field, len(fields))
	for i, f := range fields {
		zapFields[i] = zap.Any(f.Key, f.Value)
	}
	return zapFields
}
