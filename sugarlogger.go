package log

import (
	"go.uber.org/zap"
)

func DPanic(args ...interface{}) {
	defaultSugarLogger.DPanic(args...)
}

func DPanicf(template string, args ...interface{}) {
	defaultSugarLogger.DPanicf(template, args...)
}

func DPanicw(msg string, keysAndValues ...interface{}) {
	defaultSugarLogger.DPanicw(msg, keysAndValues...)
}

func Debug(args ...interface{}) {
	defaultSugarLogger.Debug(args...)
}

func Debugf(template string, args ...interface{}) {
	defaultSugarLogger.Debugf(template, args...)
}

func Debugw(msg string, keysAndValues ...interface{}) {
	defaultSugarLogger.Debugw(msg, keysAndValues...)
}

func Error(args ...interface{}) {
	defaultSugarLogger.Error(args...)
}

func Errorf(template string, args ...interface{}) {
	defaultSugarLogger.Errorf(template, args...)
}

func Errorw(msg string, keysAndValues ...interface{}) {
	defaultSugarLogger.Errorw(msg, keysAndValues...)
}

func Fatal(args ...interface{}) {
	defaultSugarLogger.Fatal(args...)
}

func Fatalf(template string, args ...interface{}) {
	defaultSugarLogger.Fatalf(template, args...)
}

func Fatalw(msg string, keysAndValues ...interface{}) {
	defaultSugarLogger.Fatalw(msg, keysAndValues...)
}

func Info(args ...interface{}) {
	defaultSugarLogger.Info(args...)
}

func Infof(template string, args ...interface{}) {
	defaultSugarLogger.Infof(template, args...)
}

func Infow(msg string, keysAndValues ...interface{}) {
	defaultSugarLogger.Infow(msg, keysAndValues...)
}

func Named(name string) *zap.SugaredLogger {
	return defaultSugarLogger.Named(name)
}

func Panic(args ...interface{}) {
	defaultSugarLogger.Panic(args...)
}

func Panicf(template string, args ...interface{}) {
	defaultSugarLogger.Panicf(template, args...)
}

func Panicw(msg string, keysAndValues ...interface{}) {
	defaultSugarLogger.Panicw(msg, keysAndValues...)
}

func Sync() error {
	return defaultSugarLogger.Sync()
}

func Warn(args ...interface{}) {
	defaultSugarLogger.Warn(args...)
}

func Warnf(template string, args ...interface{}) {
	defaultSugarLogger.Warnf(template, args...)
}

func Warnw(msg string, keysAndValues ...interface{}) {
	defaultSugarLogger.Warnw(msg, keysAndValues...)
}

func With(args ...interface{}) *zap.SugaredLogger {
	return defaultSugarLogger.With(args...)
}
