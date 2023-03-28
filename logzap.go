package logzap

import (
	"os"
	"strings"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	_global = New(Config{Level: zap.DebugLevel})
)

func Get(name string) Logger {
	return _global.Get(name)
}

func Use(core zapcore.Core) error {
	return _global.Use(core)
}

func Reload(c Config) error {
	return _global.Reload(c)
}

func New(
	c Config,
) *Logzap {
	t := &Logzap{
		level:   c.Level,
		modules: map[string]*logger{},
		log:     zap.New(zapcore.NewTee(load(c.Level)...)),
	}

	return t
}

type Logzap struct {
	level   zapcore.Level
	mu      sync.RWMutex
	log     *zap.Logger
	modules map[string]*logger
}

func load(lv zapcore.Level) []zapcore.Core {
	cores := []zapcore.Core{}
	encoder := &filterEncoder{zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())}
	if lv < zapcore.ErrorLevel {
		cores = append(cores, zapcore.NewCore(
			encoder,
			zapcore.Lock(os.Stdout),
			zap.LevelEnablerFunc(func(level zapcore.Level) bool {
				return level >= lv && lv < zapcore.ErrorLevel
			}),
		))
	}
	cores = append(cores, zapcore.NewCore(
		encoder,
		zapcore.Lock(os.Stderr),
		zap.LevelEnablerFunc(func(level zapcore.Level) bool {
			return level > lv && level >= zapcore.ErrorLevel
		}),
	))

	return cores
}

func (t *Logzap) Get(name string, opts ...zap.Option) Logger {
	t.mu.RLock()
	defer t.mu.RUnlock()
	l, ok := t.modules[name]
	if !ok {
		l = &logger{
			level: t.level,
		}
		t.modules[name] = l
	}
	l.opts = opts
	l.instance.Store(t.log.Named(name).Sugar().WithOptions(opts...))

	return l
}

func (t *Logzap) reload(cores ...zapcore.Core) error {
	// update instance
	t.log = zap.New(zapcore.NewTee(cores...))
	for name, l := range t.modules {
		l.instance.Store(t.log.Named(name).Sugar())
	}

	return nil
}

func (t *Logzap) Reload(c Config) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.level = c.Level
	for name, lv := range c.Modules {
		name = strings.ToLower(name)
		l, ok := t.modules[name]
		if !ok {
			l = &logger{}
			t.modules[name] = l
		}
		l.level = lv
	}

	return t.reload(load(c.Level)...)
}

func (t *Logzap) Use(core zapcore.Core) error {
	cores := load(t.level)
	cores = append(cores, core)
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.reload(cores...)
}
