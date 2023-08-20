package logzap

import (
	"sync/atomic"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	_global atomic.Pointer[Logzap]
)

func init() {
	_global.Store(Default())
}

// Get return a logger with given name and options from global instance.
func Get(name string, opts ...zap.Option) Logger {
	return _global.Load().Get(name, opts...)
}

// Reload reload global instance with given log level.
func Reload(lv zapcore.Level, modules ModulesLevel) {
	_global.Load().Reload(lv, modules)
}

// Sync flushes any buffered log entries.
func Sync() error {
	return _global.Load().Sync()
}
