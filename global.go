package logzap

import (
	"context"
	"sync"

	"go.uber.org/zap/zapcore"
)

var (
	_gmu    sync.Mutex
	_global = New(context.Background(), nil, Config{Middleware: Middleware{"console": "stdout"}})
)

func Get(name string) Logger {
	_gmu.Lock()
	defer _gmu.Unlock()
	return _global.Get(name)
}

func Use(m map[string]zapcore.Core) error {
	_gmu.Lock()
	defer _gmu.Unlock()
	return _global.Use(m)
}

func Reload(lv zapcore.Level, modules ModulesLevel) {
	_gmu.Lock()
	defer _gmu.Unlock()
	_global.Reload(lv, modules)
}

func Sync() error {
	_gmu.Lock()
	defer _gmu.Unlock()
	return _global.Sync()
}
