package logzap

import (
	"context"
	"errors"
	"sync/atomic"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type logger struct {
	zap.AtomicLevel
	instance atomic.Pointer[zap.SugaredLogger]
	opts     []zap.Option
}

type Logger interface {
	// Trace return true only if err != nil and log level is higher or equal error
	Trace(error) bool
	// Trace return true only if err != nil and log level is higher or equal error and context is not canceled
	TraceContext(context.Context, error) bool
	TraceReturn(error) error
	TraceReturnContext(context.Context, error) error
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
	Enabled(zapcore.Level) bool
}

func newLogger(lv zapcore.Level, log *zap.Logger, opts ...zap.Option) *logger {
	t := &logger{
		AtomicLevel: zap.NewAtomicLevelAt(lv),
		opts:        opts,
	}
	t.instance.Store(log.WithOptions(append(t.opts, zap.IncreaseLevel(lv), zap.AddCallerSkip(1))...).Sugar())
	return t
}

// reload doesn't change original options
func (t *logger) reload(lv zapcore.Level, log *zap.Logger) {
	t.AtomicLevel.SetLevel(lv)
	t.instance.Store(log.WithOptions(append(t.opts, zap.IncreaseLevel(lv), zap.AddCallerSkip(1))...).Sugar())
}

// Trace return true only if err != nil and log level is higher or equal error
func (t *logger) Trace(err error) bool {
	if err == nil {
		return false
	}
	t.instance.Load().WithOptions(zap.AddCallerSkip(1)).Error(err)

	return t.Enabled(zap.ErrorLevel)
}

// Trace return true only if err != nil and log level is higher or equal error and context is not canceled
func (t *logger) TraceContext(ctx context.Context, err error) bool {
	if err == nil {
		return false
	}

	select {
	case <-ctx.Done():
		return false
	default:
	}
	if errors.Is(err, context.Canceled) {
		panic(err)
	}
	t.instance.Load().WithOptions(zap.AddCallerSkip(1)).Error(err)

	return t.Enabled(zap.ErrorLevel)
}

func (t *logger) TraceReturn(err error) error {
	if err == nil {
		return nil
	}
	t.instance.Load().WithOptions(zap.AddCallerSkip(1)).Error(err)

	return err
}

// TraceReturnContext always return nil if context canceled
func (t *logger) TraceReturnContext(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}
	select {
	case <-ctx.Done():
		return nil
	default:
	}

	if errors.Is(err, context.Canceled) {
		panic(err)
	}
	t.instance.Load().WithOptions(zap.AddCallerSkip(1)).Error(err)

	return err
}

func (t *logger) Fatal(args ...interface{}) {
	t.instance.Load().Fatal(args...)
}

func (t *logger) Fatalf(format string, args ...interface{}) {
	t.instance.Load().Fatalf(format, args...)
}

func (t *logger) Error(args ...interface{}) {
	t.instance.Load().Error(args...)
}

func (t *logger) Errorf(format string, args ...interface{}) {
	t.instance.Load().Errorf(format, args...)
}

func (t *logger) Warn(args ...interface{}) {
	t.instance.Load().Warn(args...)
}

func (t *logger) Warnf(format string, args ...interface{}) {
	t.instance.Load().Warnf(format, args...)
}

func (t *logger) Info(args ...interface{}) {
	t.instance.Load().Info(args...)
}

func (t *logger) Infof(format string, args ...interface{}) {
	t.instance.Load().Infof(format, args...)
}

func (t *logger) Debug(args ...interface{}) {
	t.instance.Load().Debug(args...)
}

func (t *logger) Debugf(format string, args ...interface{}) {
	t.instance.Load().Debugf(format, args...)
}
