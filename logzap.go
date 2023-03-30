package logzap

import (
	"context"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var _global = &Logzap{
	log:      zap.NewNop(),
	modules:  map[string]*logger{},
	shutdown: map[string]Shutdown{},
	cores:    map[string]zapcore.Core{},
}

func Get(name string) Logger {
	return _global.Get(name)
}

func Use(m map[string]zapcore.Core) error {
	return _global.Use(m)
}

func Reload(lv zapcore.Level, modules ModulesLevel) {
	_global.Reload(lv, modules)
}

func New(
	ctx context.Context,
	registry prometheus.Registerer,
	c Config,
) *Logzap {
	t, err := NewE(ctx, registry, c)
	if err != nil {
		panic(err)
	}
	return t
}

func NewE(
	ctx context.Context,
	registry prometheus.Registerer,
	c Config,
) (t *Logzap, err error) {
	t = &Logzap{
		level: c.Level,
	}
	var core zapcore.Core
	switch {
	case len(c.Middleware) == 0:
		t.cores = map[string]zapcore.Core{}
	case registry == nil:
		core, _, err = c.Middleware.BuildByName(ctx, nil, "console")
		t.cores = map[string]zapcore.Core{
			"console": core,
		}
	default:
		t.cores, t.shutdown, err = c.Middleware.Build(ctx, registry)
	}
	if err != nil {
		return nil, err
	}
	if t.log, err = newZapLogger(t.cores); err != nil {
		return nil, multierr.Append(err, t.Shutdown(ctx))
	}
	t.modules = c.Modules.build(t.log)

	return t, nil
}

func newZapLogger(m map[string]zapcore.Core) (*zap.Logger, error) {
	sink, _, err := zap.Open("stderr")
	if err != nil {
		return nil, err
	}
	cores := make([]zapcore.Core, 0, len(m))
	for _, v := range m {
		cores = append(cores, v)
	}

	return zap.New(
		zapcore.NewTee(cores...),
		zap.ErrorOutput(sink),
		zap.Development(),
		zap.AddStacktrace(zap.WarnLevel),
		zap.AddCaller(),
	), nil
}

type Logzap struct {
	mu       sync.RWMutex
	level    zapcore.Level
	log      *zap.Logger
	modules  map[string]*logger
	cores    map[string]zapcore.Core
	shutdown map[string]Shutdown
}

func (t *Logzap) Shutdown(ctx context.Context) error {
	t.mu.RLock()
	defer t.mu.RUnlock()
	err := []error{}
	for _, c := range t.shutdown {
		if e := c(ctx); e != nil {
			err = append(err, e)
		}
	}
	return multierr.Combine(err...)
}

func (t *Logzap) Get(name string, opts ...zap.Option) Logger {
	t.mu.RLock()
	defer t.mu.RUnlock()
	l, ok := t.modules[name]
	if !ok {
		l = newLogger(t.level, t.log.Named(name), opts...)
		t.modules[name] = l
	} else if len(opts) > 0 {
		l = newLogger(l.Level(), t.log.Named(name), opts...)
	}

	return l
}

// only reconfiguration the submodule level
func (t *Logzap) Reload(lv zapcore.Level, modules ModulesLevel) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.level = lv
	for name, lv := range modules {
		m, ok := t.modules[name]
		if !ok {
			t.modules[name] = newLogger(lv, t.log.Named(name))
		} else {
			m.reload(lv, t.log.Named(name))
		}
	}
	for name, m := range t.modules {
		// reset the submodules to new general level which not in new Modules configuration
		if _, ok := modules[name]; !ok {
			m.reload(t.level, t.log.Named(name))
		}
	}
}

func (t *Logzap) Use(m map[string]zapcore.Core) (err error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	for k, c := range m {
		if _, ok := t.cores[k]; ok {
			t.log.Sugar().Warnf("middleware %s exist. ignore.", k)
			continue
		}
		t.cores[k] = c
	}
	if t.log, err = newZapLogger(t.cores); err != nil {
		return
	}
	for _, m := range t.modules {
		m.reload(m.Level(), t.log)
	}

	return
}
