package logger

import "testing"

func TestNewLogger_AndMethods_DoNotPanic(t *testing.T) {
	t.Parallel()

	logger := NewLogger()
	if logger == nil {
		t.Fatal("expected non-nil logger")
	}

	logger.Info("info", "key", "value")
	logger.Warn("warn", "key", "value")
	logger.Error("error", "key", "value")
}
