package services

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ServiceLogger provides structured logging for services
type ServiceLogger struct {
	logger *zap.Logger
	sugar  *zap.SugaredLogger
}

// LogLevel represents different logging levels
type LogLevel string

const (
	LogLevelDebug   LogLevel = "debug"
	LogLevelInfo    LogLevel = "info"
	LogLevelWarning LogLevel = "warning"
	LogLevelError   LogLevel = "error"
	LogLevelFatal   LogLevel = "fatal"
)

// LogEvent represents a structured log event
type LogEvent struct {
	Service     string                 `json:"service"`
	Method      string                 `json:"method"`
	Level       LogLevel               `json:"level"`
	Message     string                 `json:"message"`
	Error       error                  `json:"error,omitempty"`
	Duration    time.Duration          `json:"duration,omitempty"`
	UserID      uint                   `json:"user_id,omitempty"`
	EntityID    uint                   `json:"entity_id,omitempty"`
	EntityType  string                 `json:"entity_type,omitempty"`
	CacheHit    bool                   `json:"cache_hit,omitempty"`
	CacheKey    string                 `json:"cache_key,omitempty"`
	DatabaseOps int                    `json:"database_ops,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	TraceID     string                 `json:"trace_id,omitempty"`
}

// NewServiceLogger creates a new service logger
func NewServiceLogger(level LogLevel, serviceName string) *ServiceLogger {
	// Create encoder config
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	// Parse log level
	var zapLevel zapcore.Level
	switch level {
	case LogLevelDebug:
		zapLevel = zapcore.DebugLevel
	case LogLevelInfo:
		zapLevel = zapcore.InfoLevel
	case LogLevelWarning:
		zapLevel = zapcore.WarnLevel
	case LogLevelError:
		zapLevel = zapcore.ErrorLevel
	case LogLevelFatal:
		zapLevel = zapcore.FatalLevel
	default:
		zapLevel = zapcore.InfoLevel
	}

	// Create logger
	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(zapLevel)
	config.EncoderConfig = encoderConfig
	config.OutputPaths = []string{"stdout", "logs/services.log"}
	config.ErrorOutputPaths = []string{"stderr", "logs/services-error.log"}

	// Ensure log directory exists
	if err := os.MkdirAll("logs", 0755); err != nil {
		log.Printf("Warning: failed to create logs directory: %v", err)
		config.OutputPaths = []string{"stdout"}
		config.ErrorOutputPaths = []string{"stderr"}
	}

	logger, err := config.Build()
	if err != nil {
		log.Printf("Warning: failed to create structured logger, falling back to standard logger: %v", err)
		return &ServiceLogger{
			logger: zap.NewNop(),
			sugar:  zap.NewNop().Sugar(),
		}
	}

	return &ServiceLogger{
		logger: logger.Named(serviceName),
		sugar:  logger.Named(serviceName).Sugar(),
	}
}

// Debug logs a debug message
func (l *ServiceLogger) Debug(ctx context.Context, method, message string, fields ...map[string]interface{}) {
	l.log(ctx, LogLevelDebug, method, message, nil, fields...)
}

// Info logs an info message
func (l *ServiceLogger) Info(ctx context.Context, method, message string, fields ...map[string]interface{}) {
	l.log(ctx, LogLevelInfo, method, message, nil, fields...)
}

// Warning logs a warning message
func (l *ServiceLogger) Warning(ctx context.Context, method, message string, fields ...map[string]interface{}) {
	l.log(ctx, LogLevelWarning, method, message, nil, fields...)
}

// Error logs an error message
func (l *ServiceLogger) Error(ctx context.Context, method, message string, err error, fields ...map[string]interface{}) {
	l.log(ctx, LogLevelError, method, message, err, fields...)
}

// Fatal logs a fatal message and exits
func (l *ServiceLogger) Fatal(ctx context.Context, method, message string, err error, fields ...map[string]interface{}) {
	l.log(ctx, LogLevelFatal, method, message, err, fields...)
	os.Exit(1)
}

// Performance logs performance metrics
func (l *ServiceLogger) Performance(ctx context.Context, method string, duration time.Duration, cacheHit bool, dbOps int, fields ...map[string]interface{}) {
	metadata := map[string]interface{}{
		"duration":     duration.String(),
		"cache_hit":    cacheHit,
		"database_ops": dbOps,
	}

	// Merge additional fields
	for _, field := range fields {
		for k, v := range field {
			metadata[k] = v
		}
	}

	l.Info(ctx, method, "Performance metrics", metadata)
}

// Cache logs cache operations
func (l *ServiceLogger) Cache(ctx context.Context, method, operation, cacheKey string, success bool, fields ...map[string]interface{}) {
	metadata := map[string]interface{}{
		"operation": operation,
		"cache_key": cacheKey,
		"success":   success,
	}

	// Merge additional fields
	for _, field := range fields {
		for k, v := range field {
			metadata[k] = v
		}
	}

	level := LogLevelInfo
	if !success {
		level = LogLevelWarning
	}

	l.log(ctx, level, method, "Cache operation", nil, metadata)
}

// Database logs database operations
func (l *ServiceLogger) Database(ctx context.Context, method, operation string, table string, duration time.Duration, rowsAffected int64, err error, fields ...map[string]interface{}) {
	metadata := map[string]interface{}{
		"operation":     operation,
		"table":         table,
		"duration":      duration.String(),
		"rows_affected": rowsAffected,
	}

	// Merge additional fields
	for _, field := range fields {
		for k, v := range field {
			metadata[k] = v
		}
	}

	level := LogLevelInfo
	if err != nil {
		level = LogLevelError
	}

	l.log(ctx, level, method, "Database operation", err, metadata)
}

// Security logs security-related events
func (l *ServiceLogger) Security(ctx context.Context, method, event string, userID uint, success bool, fields ...map[string]interface{}) {
	metadata := map[string]interface{}{
		"event":   event,
		"user_id": userID,
		"success": success,
	}

	// Merge additional fields
	for _, field := range fields {
		for k, v := range field {
			metadata[k] = v
		}
	}

	level := LogLevelInfo
	if !success {
		level = LogLevelWarning
	}

	l.log(ctx, level, method, "Security event", nil, metadata)
}

// Audit logs audit trail events
func (l *ServiceLogger) Audit(ctx context.Context, method, action string, userID uint, entityID uint, entityType string, fields ...map[string]interface{}) {
	metadata := map[string]interface{}{
		"action":      action,
		"user_id":     userID,
		"entity_id":   entityID,
		"entity_type": entityType,
	}

	// Merge additional fields
	for _, field := range fields {
		for k, v := range field {
			metadata[k] = v
		}
	}

	l.Info(ctx, method, "Audit trail", metadata)
}

// log is the internal logging method
func (l *ServiceLogger) log(ctx context.Context, level LogLevel, method, message string, err error, fields ...map[string]interface{}) {
	// Get caller information
	_, file, line, ok := runtime.Caller(2)
	caller := "unknown"
	if ok {
		caller = fmt.Sprintf("%s:%d", file, line)
	}

	// Extract trace ID from context if available
	traceID := ""
	if ctx != nil {
		if id, ok := ctx.Value("trace_id").(string); ok {
			traceID = id
		}
	}

	// Create log event
	event := LogEvent{
		Service:   l.logger.Name(),
		Method:    method,
		Level:     level,
		Message:   message,
		Error:     err,
		Timestamp: time.Now(),
		TraceID:   traceID,
	}

	// Merge fields into metadata
	if len(fields) > 0 {
		event.Metadata = make(map[string]interface{})
		for _, field := range fields {
			for k, v := range field {
				event.Metadata[k] = v
			}
		}
	}

	// Convert to zap fields
	zapFields := []zap.Field{
		zap.String("service", event.Service),
		zap.String("method", event.Method),
		zap.String("level", string(event.Level)),
		zap.String("message", event.Message),
		zap.Time("timestamp", event.Timestamp),
		zap.String("caller", caller),
	}

	if event.Error != nil {
		zapFields = append(zapFields, zap.Error(event.Error))
	}

	if event.TraceID != "" {
		zapFields = append(zapFields, zap.String("trace_id", event.TraceID))
	}

	if event.Metadata != nil {
		zapFields = append(zapFields, zap.Any("metadata", event.Metadata))
	}

	// Log based on level
	switch level {
	case LogLevelDebug:
		l.logger.Debug(message, zapFields...)
	case LogLevelInfo:
		l.logger.Info(message, zapFields...)
	case LogLevelWarning:
		l.logger.Warn(message, zapFields...)
	case LogLevelError:
		l.logger.Error(message, zapFields...)
	case LogLevelFatal:
		l.logger.Fatal(message, zapFields...)
	}
}

// Sync flushes any buffered log entries
func (l *ServiceLogger) Sync() error {
	return l.logger.Sync()
}

// WithContext creates a new logger with context-specific fields
func (l *ServiceLogger) WithContext(ctx context.Context) *ServiceLogger {
	if ctx == nil {
		return l
	}

	// Extract common context fields
	fields := make(map[string]interface{})

	if userID, ok := ctx.Value("user_id").(uint); ok {
		fields["user_id"] = userID
	}

	if requestID, ok := ctx.Value("request_id").(string); ok {
		fields["request_id"] = requestID
	}

	if traceID, ok := ctx.Value("trace_id").(string); ok {
		fields["trace_id"] = traceID
	}

	// Create a new logger with context fields
	newLogger := &ServiceLogger{
		logger: l.logger,
		sugar:  l.sugar,
	}

	return newLogger
}

// Global logger instance
var GlobalLogger *ServiceLogger

// InitializeLogger initializes the global logger
func InitializeLogger(level LogLevel, serviceName string) {
	GlobalLogger = NewServiceLogger(level, serviceName)
}

// GetLogger returns the global logger or creates a new one
func GetLogger(serviceName string) *ServiceLogger {
	if GlobalLogger != nil {
		return GlobalLogger
	}
	return NewServiceLogger(LogLevelInfo, serviceName)
}
