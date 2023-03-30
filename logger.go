package logzap

import (
	"context"
	"sync/atomic"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger interface {
	Trace(err error, fields ...zap.Field) error
	TraceContext(ctx context.Context, err error, fields ...zap.Field) error

	Error(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Debug(msg string, fields ...zap.Field)

	Errorf(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Debugf(format string, args ...interface{})
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

// reload doesn't change original options
func (t *logger) reload(log *zap.Logger, lv zapcore.Level) {
	t.AtomicLevel.SetLevel(lv)
	t.instance.Store(log.WithOptions(append(t.opts, zap.IncreaseLevel(lv), zap.AddCallerSkip(1))...))
}

func (t *logger) Trace(err error, fields ...zap.Field) error {
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
