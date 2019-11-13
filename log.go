package log

import (
	"go.uber.org/zap"
)

var (
	defaultLogger      *zap.Logger
	defaultSugarLogger *zap.SugaredLogger
)

func init() {
	setDefaultLogger()
}

func setDefaultLogger() {
	logger, err := NewDevelopmentConfig().Build()
	if err != nil {
		panic(err) // this should not happen, if it does, we need to fix it
	}

	logger = logger.WithOptions(
		zap.AddCaller(),
		zap.AddCallerSkip(1),
	)

	Use(logger)
}

// Use sets the default logger used by the package
func Use(logger *zap.Logger) {
	defaultLogger = logger
	defaultSugarLogger = logger.Sugar()
}

// CapturePanic captures, logs and re-throws a panic.
// Only useful when used with defer.
func CapturePanic() {
	if r := recover(); r != nil {
		l := defaultLogger.WithOptions(zap.AddCallerSkip(1))
		l.Sugar().Panic(r)
	}
}
