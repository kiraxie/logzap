package logzap

import (
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	_global, err = New(Config{Level: zap.DebugLevel})
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

func New(c Config) (*Logzap, error) {
	l, err := newZapLogger([]string{"stdout"}, nil)
	if err != nil {
		return nil, err
	}
	t := &Logzap{
		config:  c,
		log:     l,
		modules: c.Modules.build(l),
	}

	return t, nil
}

func newZapLogger(line []string, json []string, cores ...zapcore.Core) (*zap.Logger, error) {
	encoderConf := zap.NewDevelopmentEncoderConfig()
	lineEnc := &filterEncoder{zapcore.NewConsoleEncoder(encoderConf)}
	jsonEnc := &filterEncoder{zapcore.NewJSONEncoder(encoderConf)}
	errSink, _, err := zap.Open("stderr")
	if err != nil {
		return nil, err
	}

	lineSink, close, err := zap.Open(line...)
	if err != nil {
		return nil, err
	}
	lineCore := zapcore.NewCore(lineEnc, lineSink, zap.DebugLevel)

	jsonSink, _, err := zap.Open(json...)
	if err != nil {
		close()
		return nil, err
	}
	jsonCore := zapcore.NewCore(jsonEnc, jsonSink, zap.DebugLevel)

	return zap.New(
		zapcore.NewTee(append([]zapcore.Core{lineCore, jsonCore}, cores...)...),
		zap.ErrorOutput(errSink),
		zap.Development(),
		zap.AddStacktrace(zap.WarnLevel),
		zap.AddCaller(),
	), nil
}

type Logzap struct {
	mu      sync.RWMutex
	log     *zap.Logger
	modules map[string]*logger

	config Config
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
		l = newLogger(t.config.Level, t.log.Named(name), opts...)
		t.modules[name] = l
	} else if len(opts) > 0 {
		l = newLogger(l.Level(), t.log.Named(name), opts...)
	}

	return l
}

func (t *Logzap) reload(c Config) error {
	t.config = c
	for name, lv := range c.Modules {
		m, ok := t.modules[name]
		if !ok {
			t.modules[name] = newLogger(lv, t.log.Named(name))
		} else {
			m.reload(lv, t.log.Named(name))
		}
	}
	for name, m := range t.modules {
		// reset the submodules to new general level which not in new Modules configuration
		if _, ok := c.Modules[name]; !ok {
			m.reload(t.config.Level, t.log.Named(name))
		}
	}

	return nil
}

func (t *Logzap) Reload(c Config) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.reload(c)
}

func (t *Logzap) Use(core zapcore.Core) (err error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.log, err = newZapLogger([]string{"stdout"}, nil, core); err != nil {
		return err
	}

	return t.reload(t.config)
}
