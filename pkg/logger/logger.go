package logger

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/nikita-belan-01/netcraze_route_builder/pkg/strutil"
)

type loggerKey string

const LoggerKeyName loggerKey = "log"

const maxLoggedError = 4096

const (
	FormatText = "text"
	FormatJSON = "json"
)

var (
	ErrInvalidLogLevel  = errors.New("invalid log level")
	ErrInvalidLogFormat = errors.New("invalid log format")
)

type Logger struct {
	*slog.Logger
	file *os.File
}

func ErrorAttr(err error) slog.Attr {
	if err == nil {
		return slog.String("error", "")
	}

	return slog.String("error", strutil.Shorten(err.Error(), maxLoggedError))
}

func ParseLevel(logLevel string) (slog.Level, error) {
	switch strings.ToLower(strings.TrimSpace(logLevel)) {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return 0, fmt.Errorf("%w: %q", ErrInvalidLogLevel, logLevel)
	}
}

func ValidateLogLevel(logLevel string) error {
	if _, err := ParseLevel(logLevel); err != nil {
		return err
	}

	return nil
}

func ValidateLogFormat(logFormat string) error {
	switch logFormat {
	case FormatText, FormatJSON:
		return nil
	default:
		return fmt.Errorf("%w: %q", ErrInvalidLogFormat, logFormat)
	}
}

func New(logDir, logLevel, logFormat string) (*Logger, error) {
	level, err := ParseLevel(logLevel)
	if err != nil {
		return nil, err
	}

	if err := ValidateLogFormat(logFormat); err != nil {
		return nil, err
	}

	var (
		writers = []io.Writer{os.Stderr}
		logFile *os.File
	)

	if logDir != "" {
		if err := os.MkdirAll(logDir, 0o755); err != nil {
			return nil, fmt.Errorf("create log dir: %w", err)
		}

		nowStr := time.Now().UTC().Format("2006-01-02T15-04-05.000000")
		logFilePath := filepath.Join(logDir, fmt.Sprintf("%s.log", nowStr))

		logFile, err = os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			return nil, fmt.Errorf("create log file: %w", err)
		}

		writers = append(writers, logFile)
	}

	options := &slog.HandlerOptions{Level: level}

	var handler slog.Handler
	output := io.MultiWriter(writers...)

	if logFormat == FormatJSON {
		handler = slog.NewJSONHandler(output, options)
	} else {
		handler = slog.NewTextHandler(output, options)
	}

	return &Logger{
		Logger: slog.New(handler),
		file:   logFile,
	}, nil
}

func Discard() *Logger {
	return &Logger{
		Logger: slog.New(slog.DiscardHandler),
	}
}

func (l *Logger) With(args ...any) *Logger {
	return &Logger{
		Logger: l.Logger.With(args...),
		file:   l.file,
	}
}

func (l *Logger) Close() error {
	if l.file == nil {
		return nil
	}

	var fileSyncErr, fileCloseErr error

	if err := l.file.Sync(); err != nil {
		fileSyncErr = fmt.Errorf("sync logger file: %w", err)
	}

	if err := l.file.Close(); err != nil {
		fileCloseErr = fmt.Errorf("close logger file: %w", err)
	}

	return errors.Join(fileSyncErr, fileCloseErr)
}

func NewContext(ctx context.Context, log *Logger) context.Context {
	return context.WithValue(ctx, LoggerKeyName, log)
}

func FromContext(ctx context.Context) *Logger {
	log, ok := ctx.Value(LoggerKeyName).(*Logger)
	if !ok {
		return Discard()
	}

	return log
}
