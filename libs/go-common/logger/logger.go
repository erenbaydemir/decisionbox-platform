package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger wraps zap.Logger with convenience methods.
type Logger struct {
	*zap.Logger
}

// New creates a structured JSON logger for production use.
// level: "debug", "info", "warn", "error"
func New(service, level string) *Logger {
	lvl := parseLevel(level)

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "severity",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.MillisDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// JSON for production, console for dev
	var encoder zapcore.Encoder
	env := os.Getenv("ENV")
	if env == "prod" || env == "production" {
		encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	core := zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), lvl)
	l := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	l = l.With(zap.String("service", service))

	return &Logger{Logger: l}
}

// With creates a child logger with additional fields.
func (l *Logger) With(fields ...zap.Field) *Logger {
	return &Logger{Logger: l.Logger.With(fields...)}
}

// Common field constructors for convenience.
func AppID(id string) zap.Field        { return zap.String("app_id", id) }
func OrgID(id string) zap.Field        { return zap.String("org_id", id) }
func UserID(id string) zap.Field       { return zap.String("user_id", id) }
func SessionID(id string) zap.Field    { return zap.String("session_id", id) }
func CorrelationID(id string) zap.Field { return zap.String("correlation_id", id) }
func Err(err error) zap.Field          { return zap.Error(err) }

func parseLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "warn", "warning":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}
