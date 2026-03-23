package logger

import (
	"log/slog"
	"os"

	"github.com/avito-internships/test-backend-1-EdOoO21/internal/ports"
)

type Logger struct {
	logger *slog.Logger
}

func NewLogger() ports.Logger {
	return &Logger{
		logger: slog.New(slog.NewTextHandler(os.Stdout, nil)),
	}
}

func (l *Logger) Info(msg string, args ...any) {
	l.logger.Info(msg, args...)
}

func (l *Logger) Warn(msg string, args ...any) {
	l.logger.Warn(msg, args...)
}

func (l *Logger) Error(msg string, args ...any) {
	l.logger.Error(msg, args...)
}
