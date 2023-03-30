package logzap

import (
	"context"
	"fmt"
	"sync"

	"github.com/kiraxie/logzap/middleware/console"
	"github.com/kiraxie/logzap/middleware/loki"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/multierr"
	"go.uber.org/zap/zapcore"
)

type Shutdown = func(context.Context) error

type MiddlewareConstructor func(ctx context.Context, registry prometheus.Registerer, url string) (zapcore.Core, Shutdown, error)

var (
	mu                     sync.RWMutex
	_middlewareConstructor = map[string]MiddlewareConstructor{
		"console": console.New,
		"loki":    loki.New,
	}
)

type Middleware map[string]string

func (t Middleware) MustBuild(
	ctx context.Context,
	registry prometheus.Registerer,
) (map[string]zapcore.Core, map[string]Shutdown) {
	middleware, close, err := t.Build(ctx, registry)
	if err != nil {
		panic(err)
	}
	return middleware, close
}

func (t Middleware) Build(
	ctx context.Context,
	registry prometheus.Registerer,
) (middleware map[string]zapcore.Core, shutdown map[string]Shutdown, err error) {
	mu.RLock()
	defer mu.RUnlock()
	close := func(ctx context.Context) error {
		var e []error
		for _, c := range shutdown {
			if err := c(ctx); err != nil {
				e = append(e, err)
			}
		}
		return multierr.Combine(e...)
	}
	for name, url := range t {
		constructor, ok := _middlewareConstructor[name]
		if !ok {
			return nil, nil, multierr.Append(
				fmt.Errorf("unsupported middleware: %s", name),
				close(ctx),
			)
		}
		m, c, err := constructor(ctx, registry, url)
		if err != nil {
			return nil, nil, multierr.Append(err, close(ctx))
		}
		middleware[name] = m
		shutdown[name] = c
	}
	return
}

func (t Middleware) BuildByName(
	ctx context.Context,
	registry prometheus.Registerer,
	name string,
) (core zapcore.Core, close Shutdown, err error) {
	mu.RLock()
	defer mu.RUnlock()

	url, ok := t[name]
	if !ok {
		return nil, nil, fmt.Errorf("middleware %s not found", name)
	}
	constructor, ok := _middlewareConstructor[name]
	if !ok {
		return nil, nil, fmt.Errorf("unsupported middleware: %s", name)
	}
	core, close, err = constructor(ctx, registry, url)
	if err != nil {
		return nil, nil, err
	}

	return
}
