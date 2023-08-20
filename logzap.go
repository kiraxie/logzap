package logzap

import (
	"context"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Nop return a instance which does nothing
func Nop() *Logzap {
	return &Logzap{
		log:     zap.NewNop(),
		modules: map[string]*logger{},
	}
}

// Default return a instance which with DebugLevel
func Default() *Logzap {
	return New(context.Background(), prometheus.DefaultRegisterer, Config{Level: zapcore.DebugLevel})
}

// New return a instance with given configuration.
// Note that it will panic if any error occurred.
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

// NewE return a instance and error with given configuration.
func NewE(
	ctx context.Context,
	registry prometheus.Registerer,
	c Config,
) (t *Logzap, err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if registry == nil {
		registry = prometheus.DefaultRegisterer
	}
	if len(c.Cores) == 0 {
		c.Cores = Cores{"console": "console://"}
	}
	t = &Logzap{
		level: c.Level,
	}
	cores, err := c.Cores.Build(ctx, registry)
	if err != nil {
		return nil, err
	}
	for _, c := range cores {
		t.syncs = append(t.syncs, c.Sync)
	}
	if t.log, err = newZapLogger(cores); err != nil {
		return nil, err
	}
	t.modules = c.Modules.build(t.log)

	return t, nil
}

func newZapLogger(cores []zapcore.Core) (*zap.Logger, error) {

	return zap.New(
		zapcore.NewTee(cores...),
		zap.Development(),
		zap.AddStacktrace(zap.WarnLevel),
		zap.AddCaller(),
	), nil
}

type Logzap struct {
	mu      sync.RWMutex
	level   zapcore.Level
	log     *zap.Logger
	modules map[string]*logger
	syncs   []func() error
}

func (t *Logzap) Sync() (err error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	for _, sync := range t.syncs {
		if e := sync(); e != nil {
			err = multierr.Append(err, e)
		}
	}

	return
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

// only reconfiguration the global level and submodule level
func (t *Logzap) Reload(lv zapcore.Level, modules ModulesLevel) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.level = lv

	for name, m := range t.modules {
		// reset the submodules to new general level which not in new Modules configuration
		if _, ok := modules[name]; !ok {
			m.reload(t.log.Named(name), t.level)
		} else {
			t.modules[name] = newLogger(t.log.Named(name), lv)
		}
	}
}

func (t *Logzap) Use(core zapcore.Core) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.log = t.log.WithOptions(zap.WrapCore(func(zapcore.Core) zapcore.Core {
		return core
	}))
	for name, m := range t.modules {
		m.reload(t.log.Named(name), m.Level())
	}
}
