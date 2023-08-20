package logzap

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/kiraxie/logzap/core/buffer"
	"github.com/kiraxie/logzap/core/console"
	"github.com/kiraxie/logzap/core/loki"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap/zapcore"
)

var (
	ErrUnsupportedCoreConstructor = errors.New("unsupported core constructor")
	ErrCoreNotFound               = errors.New("core not found")
)

type CoreConstructor func(ctx context.Context, registry prometheus.Registerer, url string) (zapcore.Core, error)

var (
	mu               sync.RWMutex
	_coreConstructor = map[string]CoreConstructor{
		"buffer":  buffer.New,
		"console": console.New,
		"loki":    loki.New,
	}
)

type Cores map[string]string

func (t Cores) MustBuild(
	ctx context.Context,
	registry prometheus.Registerer,
) []zapcore.Core {
	core, err := t.Build(ctx, registry)
	if err != nil {
		panic(err)
	}

	return core
}

func (t Cores) Build(
	ctx context.Context,
	registry prometheus.Registerer,
) (core []zapcore.Core, err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	mu.RLock()
	defer mu.RUnlock()
	core = []zapcore.Core{}
	for name, url := range t {
		constructor, ok := _coreConstructor[name]
		if !ok {
			return nil, fmt.Errorf("%w: %s", ErrUnsupportedCoreConstructor, name)
		}
		m, err := constructor(ctx, registry, url)
		if err != nil {
			return nil, err
		}
		core = append(core, m)
	}

	return
}

func (t Cores) BuildByName(
	ctx context.Context,
	registry prometheus.Registerer,
	name string,
) (core zapcore.Core, err error) {
	mu.RLock()
	defer mu.RUnlock()

	url, ok := t[name]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrCoreNotFound, name)
	}
	constructor, ok := _coreConstructor[name]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedCoreConstructor, name)
	}
	core, err = constructor(ctx, registry, url)
	if err != nil {
		return nil, err
	}

	return
}
