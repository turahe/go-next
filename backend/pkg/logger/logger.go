package logger

import (
	"bytes"
	"io"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/getsentry/sentry-go/logrus"
	"github.com/sirupsen/logrus"
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

	// Add Sentry hook to Logrus
	logger.AddHook(&sentrylogrus.Hook{})
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
