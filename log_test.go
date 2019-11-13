package log

import (
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

// setTestLogger sets the global defaultLogger to a test logger
// options set here should resemble default logger options
func setTestLogger() *observer.ObservedLogs {
	core, obs := observer.New(zapcore.DebugLevel)
	logger := zap.New(core).WithOptions(
		zap.AddCaller(),
		zap.AddCallerSkip(1),
	)
	Use(logger)
	return obs
}

func TestCapturePanic(t *testing.T) {
	obs := setTestLogger()

	defer func() {
		recover()
		logs := obs.TakeAll()
		if len(logs) != 1 {
			t.Fatal("expected a panic log")
		}
		if logs[0].Message != "oh no" {
			t.Errorf("expected 'oh no', got %v", logs[0].Message)
		}
		if logs[0].Caller.TrimmedPath() != "log/log_test.go:41" {
			t.Errorf("file:line doesn't match, got %v", logs[0].Caller.TrimmedPath())
		}
	}()

	defer CapturePanic()
	panic("oh no") // if line changes, update test above
}
