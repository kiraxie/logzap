package logzap

import (
	"context"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

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
	if len(c.Middleware) == 0 {
		c.Middleware = Middleware{"console": ""}
	}
	t = &Logzap{
		level: c.Level,
	}
	if t.cores, err = c.Middleware.Build(ctx, registry); err != nil {
		return nil, err
	}
	if t.log, err = newZapLogger(t.cores); err != nil {
		return nil, err
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
	mu    sync.RWMutex
	level zapcore.Level
	log   *zap.Logger
	// modules map[string]*logger
	modules map[string]*logger
	cores   map[string]zapcore.Core
}

func (t *Logzap) Sync() error {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return t.log.Sync()
}

// return the underlying cores, might be non thread-safe
func (t *Logzap) Cores() map[string]zapcore.Core {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return t.cores
}

func (t *Logzap) Core(name string) (zapcore.Core, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	c, ok := t.cores[name]

	return c, ok
}

func (t *Logzap) Get(name string, opts ...zap.Option) Logger {
	t.mu.RLock()
	defer t.mu.RUnlock()
	l, ok := t.modules[name]
	if !ok {
		l = newLogger(t.log.Named(name), t.level, opts...)
		t.modules[name] = l
	} else if len(opts) > 0 {
		l.opts = opts
		l.reload(t.log.Named(name), l.Level())
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
			t.modules[name] = newLogger(t.log.Named(name), lv)
		} else {
			m.reload(t.log.Named(name), lv)
		}
	}
	for name, m := range t.modules {
		// reset the submodules to new general level which not in new Modules configuration
		if _, ok := modules[name]; !ok {
			m.reload(t.log.Named(name), t.level)
		}
	}
}

func (t *Logzap) Use(m map[string]zapcore.Core) (err error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	for k, c := range m {
		if v, ok := t.cores[k]; ok {
			t.log.Sugar().Warnf("middleware %s exist. overwrite.", k)
			if err := v.Sync(); err != nil {
				return err
			}
		}
		t.cores[k] = c
	}
	if t.log, err = newZapLogger(t.cores); err != nil {
		return
	}
	for name, m := range t.modules {
		m.reload(t.log.Named(name), m.Level())
	}

	return
}
