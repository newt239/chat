package service

// Logger はロギング機能を提供するインターフェースです
type Logger interface {
	// Info は情報レベルのログを出力します
	Info(msg string, fields ...LogField)

	// Warn は警告レベルのログを出力します
	Warn(msg string, fields ...LogField)

	// Error はエラーレベルのログを出力します
	Error(msg string, fields ...LogField)

	// Debug はデバッグレベルのログを出力します
	Debug(msg string, fields ...LogField)
}

// LogField はログフィールドを表す構造体です
type LogField struct {
	Key   string
	Value interface{}
}
