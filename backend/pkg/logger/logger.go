package logger

import (
	"bytes"
	"io"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	sentrylogrus "github.com/getsentry/sentry-go/logrus"
	"github.com/sirupsen/logrus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger = logrus.New()

func init() {
	logger.Out = getWriter()
	logger.Level = logrus.InfoLevel
	logger.Formatter = &formatter{}

	logger.SetReportCaller(true)

	// Initialize Sentry
	err := sentry.Init(sentry.ClientOptions{
		Dsn: "", // Replace with your Sentry DSN
	})
	if err != nil {
		logger.Fatalf("sentry.Init: %s", err)
	}

	// Add Sentry hook to Logrus (for sentry-go/logrus v0.13.0+)
	hook, err := sentrylogrus.New([]logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
	}, sentry.ClientOptions{})
	if err == nil {
		logger.AddHook(hook)
	}
}

func SetLogLevel(level logrus.Level) {
	logger.Level = level
}

type Fields logrus.Fields

// Debugf logs a message at level Debug on the standard logger.
func Debugf(format string, args ...interface{}) {
	if logger.Level >= logrus.DebugLevel {
		entry := logger.WithFields(logrus.Fields{})
		entry.Debugf(format, args...)
	}
}

// Infof logs a message at level Info on the standard logger.
func Infof(format string, args ...interface{}) {
	if logger.Level >= logrus.InfoLevel {
		entry := logger.WithFields(logrus.Fields{})
		entry.Infof(format, args...)
	}
}

// Warnf logs a message at level Warn on the standard logger.
func Warnf(format string, args ...interface{}) {
	if logger.Level >= logrus.WarnLevel {
		entry := logger.WithFields(logrus.Fields{})
		entry.Warnf(format, args...)
	}
}

// Errorf logs a message at level Error on the standard logger.
func Errorf(format string, args ...interface{}) {
	if logger.Level >= logrus.ErrorLevel {
		entry := logger.WithFields(logrus.Fields{})
		entry.Errorf(format, args...)
	}
}

// Fatalf logs a message at level Fatal on the standard logger.
func Fatalf(format string, args ...interface{}) {
	if logger.Level >= logrus.FatalLevel {
		entry := logger.WithFields(logrus.Fields{})
		entry.Fatalf(format, args...)
	}
}

func getWriter() io.Writer {
	if _, err := os.Stat("./log"); os.IsNotExist(err) {
		os.MkdirAll("./log", os.ModePerm)
	}

	file, err := os.OpenFile("log/application.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		logger.Errorf("Failed to open log file: %v", err)
		return os.Stdout
	} else {
		return file
	}
}

// Formatter implements logrus.Formatter interface.
type formatter struct {
	prefix string
}

// Format building log message.
func (f *formatter) Format(entry *logrus.Entry) ([]byte, error) {
	var sb bytes.Buffer

	var newLine = "\n"
	if runtime.GOOS == "windows" {
		newLine = "\r\n"
	}

	sb.WriteString(strings.ToUpper(entry.Level.String()))
	sb.WriteString(" ")
	sb.WriteString(entry.Time.Format(time.RFC3339))
	sb.WriteString(" ")
	sb.WriteString(f.prefix)
	sb.WriteString(entry.Message)
	sb.WriteString(newLine)

	return sb.Bytes(), nil
}

// --- Structured Service Logger (zap-based) ---
// Provides structured logging for services, used by internal/services

// LogLevel represents different logging levels
// (from services/logger.go)
type LogLevel string

const (
	LogLevelDebug   LogLevel = "debug"
	LogLevelInfo    LogLevel = "info"
	LogLevelWarning LogLevel = "warning"
	LogLevelError   LogLevel = "error"
	LogLevelFatal   LogLevel = "fatal"
)

// LogEvent represents a structured log event
// (from services/logger.go)
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

// ServiceLogger provides structured logging for services
// (from services/logger.go)
type ServiceLogger struct {
	logger *zap.Logger
	sugar  *zap.SugaredLogger
}

// NewServiceLogger creates a new zap-based ServiceLogger
func NewServiceLogger(level LogLevel, serviceName string) *ServiceLogger {
	var cfg zap.Config
	if level == LogLevelDebug {
		cfg = zap.NewDevelopmentConfig()
	} else {
		cfg = zap.NewProductionConfig()
	}
	cfg.Level = zap.NewAtomicLevelAt(zapcore.Level(parseZapLevel(level)))
	cfg.OutputPaths = []string{"stdout"}
	cfg.InitialFields = map[string]interface{}{"service": serviceName}
	logger, _ := cfg.Build()
	return &ServiceLogger{
		logger: logger,
		sugar:  logger.Sugar(),
	}
}

// parseZapLevel converts LogLevel to zapcore.Level
func parseZapLevel(level LogLevel) zapcore.Level {
	switch level {
	case LogLevelDebug:
		return zapcore.DebugLevel
	case LogLevelInfo:
		return zapcore.InfoLevel
	case LogLevelWarning:
		return zapcore.WarnLevel
	case LogLevelError:
		return zapcore.ErrorLevel
	case LogLevelFatal:
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

// InitializeLogger sets up a global zap logger for the service
func InitializeLogger(level LogLevel, serviceName string) {
	logger := NewServiceLogger(level, serviceName)
	zap.ReplaceGlobals(logger.logger)
}

// Info logs an info message with optional key-value pairs
func (l *ServiceLogger) Info(msg string, keysAndValues ...interface{}) {
	l.sugar.Infow(msg, keysAndValues...)
}

// Debug logs a debug message with optional key-value pairs
func (l *ServiceLogger) Debug(msg string, keysAndValues ...interface{}) {
	l.sugar.Debugw(msg, keysAndValues...)
}

// Warn logs a warning message with optional key-value pairs
func (l *ServiceLogger) Warn(msg string, keysAndValues ...interface{}) {
	l.sugar.Warnw(msg, keysAndValues...)
}

// Error logs an error message with optional key-value pairs
func (l *ServiceLogger) Error(msg string, keysAndValues ...interface{}) {
	l.sugar.Errorw(msg, keysAndValues...)
}

// Fatal logs a fatal message with optional key-value pairs and exits
func (l *ServiceLogger) Fatal(msg string, keysAndValues ...interface{}) {
	l.sugar.Fatalw(msg, keysAndValues...)
}

// ... (copy all ServiceLogger methods, NewServiceLogger, InitializeLogger, GetLogger, etc. from services/logger.go) ...
