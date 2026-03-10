package observability

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewZapLogger creates a production-ready zap logger.
// CRITICAL: All output MUST go to stderr for MCP Stdio protocol compliance.
func NewZapLogger(logLevel string) (*zap.Logger, error) {
	level := zap.InfoLevel
	switch logLevel {
	case "debug":
		level = zap.DebugLevel
	case "warn":
		level = zap.WarnLevel
	case "error":
		level = zap.ErrorLevel
	}

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// Explicitly configure to write ONLY to stderr
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.Lock(zapcore.AddSync(os.Stderr)), // Lock ensures thread-safe stderr writes
		level,
	)

	// Disable any default stdout/stderr sampling that might leak to stdout
	return zap.New(core, zap.ErrorOutput(zapcore.Lock(zapcore.AddSync(os.Stderr)))), nil
}
