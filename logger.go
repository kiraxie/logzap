package logzap

import (
	"context"
	"sync/atomic"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger interface {
	// Error logs a message at ErrorLevel if err is not nil.
	Trace(err error, fields ...zap.Field)
	// TraceError logs a message at ErrorLevel if err is not nil.
	TraceError(err error, fields ...zap.Field) error
	// TraceContext logs a message at ErrorLevel if err is not nil and ctx is not done.
	TraceContext(ctx context.Context, err error, fields ...zap.Field) error

	// Error logs a message at ErrorLevel. The message includes any fields passed
	Error(msg string, fields ...zap.Field)
	// Warn logs a message at WarnLevel. The message includes any fields passed
	Warn(msg string, fields ...zap.Field)
	// Info logs a message at InfoLevel. The message includes any fields passed
	Info(msg string, fields ...zap.Field)
	// Debug logs a message at DebugLevel. The message includes any fields passed
	Debug(msg string, fields ...zap.Field)

	// Errorf uses fmt.Sprintf to log a templated message at ErrorLevel.
	Errorf(format string, args ...interface{})
	// Warnf uses fmt.Sprintf to log a templated message at WarnLevel.
	Warnf(format string, args ...interface{})
	// Infof uses fmt.Sprintf to log a templated message at InfoLevel.
	Infof(format string, args ...interface{})
	// Debugf uses fmt.Sprintf to log a templated message at DebugLevel.
	Debugf(format string, args ...interface{})

	// Increase increase log level
	Increase(lv zapcore.Level) Logger
	// Level return
	zapcore.LevelEnabler
	// L return zap.Logger
	L() *zap.Logger
}

func newLogger(log *zap.Logger, lv zapcore.Level, opts ...zap.Option) *logger {
	l := &logger{
		AtomicLevel: zap.NewAtomicLevelAt(lv),
		opts:        opts,
	}
	l.instance.Store(log.WithOptions(append(opts, zap.IncreaseLevel(lv), zap.AddCallerSkip(1))...))

	return l
}

type logger struct {
	zap.AtomicLevel
	instance atomic.Pointer[zap.Logger]
	opts     []zap.Option
}

func (t *logger) L() *zap.Logger {
	return t.instance.Load()
}

// reload doesn't change original options
func (t *logger) reload(log *zap.Logger, lv zapcore.Level) {
	t.AtomicLevel.SetLevel(lv)
	t.instance.Store(log.WithOptions(append(t.opts, zap.IncreaseLevel(lv), zap.AddCallerSkip(1))...))
}

func (t *logger) Increase(lv zapcore.Level) Logger {
	if !t.Level().Enabled(lv) {
		return t
	}
	l := &logger{
		AtomicLevel: zap.NewAtomicLevelAt(lv),
		opts:        t.opts,
	}
	l.instance.Store(t.instance.Load().WithOptions(zap.IncreaseLevel(lv)))

	return l
}

func (t *logger) Trace(err error, fields ...zap.Field) {
	if err == nil {
		return
	}
	t.instance.Load().Error(err.Error(), fields...)
}

func (t *logger) TraceError(err error, fields ...zap.Field) error {
	if err == nil {
		return nil
	}
	t.instance.Load().Error(err.Error(), fields...)

	return err
}

func (t *logger) TraceContext(ctx context.Context, err error, fields ...zap.Field) error {
	if err == nil {
		return nil
	}
	select {
	case <-ctx.Done():
		return nil
	default:
	}
	t.instance.Load().Error(err.Error(), fields...)

	return err
}

func (t *logger) Error(msg string, fields ...zap.Field) {
	t.instance.Load().Error(msg, fields...)
}

func (t *logger) Warn(msg string, fields ...zap.Field) {
	t.instance.Load().Warn(msg, fields...)
}

func (t *logger) Info(msg string, fields ...zap.Field) {
	t.instance.Load().Info(msg, fields...)
}

func (t *logger) Debug(msg string, fields ...zap.Field) {
	t.instance.Load().Debug(msg, fields...)
}

func (t *logger) Errorf(format string, args ...interface{}) {
	t.instance.Load().Sugar().Errorf(format, args...)
}

func (t *logger) Warnf(format string, args ...interface{}) {
	t.instance.Load().Sugar().Warnf(format, args...)
}

func (t *logger) Infof(format string, args ...interface{}) {
	t.instance.Load().Sugar().Infof(format, args...)
}

func (t *logger) Debugf(format string, args ...interface{}) {
	t.instance.Load().Sugar().Debugf(format, args...)
}
