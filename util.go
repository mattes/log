package log

import (
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Sampling is a convenience function that returns a zap.Option
// which wraps a core with a sample policy.
func Sampling(initial, thereafter int) zap.Option {
	return zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return zapcore.NewSampler(core, time.Second, initial, thereafter)
	})
}

// ErrorOutput is a convenience function that creates a zapcore.WriteSyncer
// and returns it as zap.ErrorOutput option. It panics if path can't be opened.
// Example: ErrorOutput("stderr")
func ErrorOutput(paths ...string) zap.Option {
	w, _, err := zap.Open(paths...)
	if err != nil {
		panic(err)
	}
	return zap.ErrorOutput(w)
}
