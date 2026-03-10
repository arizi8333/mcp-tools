package observability

import (
	"time"

	"go.uber.org/zap"
)

// AuditStore defines the interface for recording tool executions.
type AuditStore interface {
	Record(tool string, duration time.Duration, success bool) error
}

// Logger provides structured logging for the server.
type Logger struct {
	Zap   *zap.Logger
	Audit AuditStore
}

func (l *Logger) Info(msg string, args ...any) {
	if l.Zap != nil {
		l.Zap.Sugar().Infof(msg, args...)
	}
}

func (l *Logger) Error(msg string, err error, args ...any) {
	if l.Zap != nil {
		l.Zap.Error(msg, zap.Error(err), zap.Any("args", args))
	}
}

// LogExecution tracks tool execution details and records metrics.
func (l *Logger) LogExecution(toolName string, duration time.Duration, success bool) {
	status := "SUCCESS"
	if !success {
		status = "FAILED"
	}

	if l.Zap != nil {
		l.Zap.Info("Tool execution",
			zap.String("tool", toolName),
			zap.String("status", status),
			zap.Duration("duration", duration),
		)
	}

	// Record Prometheus Metrics
	ToolExecutionCounter.WithLabelValues(toolName, status).Inc()
	ToolExecutionDuration.WithLabelValues(toolName).Observe(duration.Seconds())

	// Record to Audit Store
	if l.Audit != nil {
		if err := l.Audit.Record(toolName, duration, success); err != nil {
			l.Error("Failed to record audit", err)
		}
	}
}

// IncrementRequestCount increments the registry request counter.
func (l *Logger) IncrementRequestCount(toolName string) {
	RequestCounter.WithLabelValues(toolName).Inc()
}
