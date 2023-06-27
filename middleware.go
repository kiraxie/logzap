package logzap

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/kiraxie/logzap/middleware/buffer"
	"github.com/kiraxie/logzap/middleware/console"
	"github.com/kiraxie/logzap/middleware/loki"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap/zapcore"
)

var (
	ErrUnsupportedMiddleware = errors.New("unsupported middleware")
	ErrMiddlewareNotFound    = errors.New("middleware not found")
)

type MiddlewareConstructor func(ctx context.Context, registry prometheus.Registerer, url string) (zapcore.Core, error)

var (
	mu                     sync.RWMutex
	_middlewareConstructor = map[string]MiddlewareConstructor{
		"buffer":  buffer.New,
		"console": console.New,
		"loki":    loki.New,
	}
)

type Middleware map[string]string

func (t Middleware) MustBuild(
	ctx context.Context,
	registry prometheus.Registerer,
) map[string]zapcore.Core {
	middleware, err := t.Build(ctx, registry)
	if err != nil {
		panic(err)
	}

	return middleware
}

func (t Middleware) Build(
	ctx context.Context,
	registry prometheus.Registerer,
) (middleware map[string]zapcore.Core, err error) {
	mu.RLock()
	defer mu.RUnlock()
	middleware = map[string]zapcore.Core{}
	for name, url := range t {
		constructor, ok := _middlewareConstructor[name]
		if !ok {
			return nil, fmt.Errorf("%w: %s", ErrUnsupportedMiddleware, name)
		}
		m, err := constructor(ctx, registry, url)
		if err != nil {
			return nil, err
		}
		middleware[name] = m
	}

	return
}

func (t Middleware) BuildByName(
	ctx context.Context,
	registry prometheus.Registerer,
	name string,
) (core zapcore.Core, err error) {
	mu.RLock()
	defer mu.RUnlock()

	url, ok := t[name]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrMiddlewareNotFound, name)
	}
	constructor, ok := _middlewareConstructor[name]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedMiddleware, name)
	}
	core, err = constructor(ctx, registry, url)
	if err != nil {
		return nil, err
	}

	return
}
