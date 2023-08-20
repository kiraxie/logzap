package console

import (
	"context"
	"net/url"
	"os"

	"github.com/kiraxie/logzap/filter"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// console://?encoder=json&filter=true
func New(
	_ context.Context,
	_ prometheus.Registerer,
	rawURL string,
) (zapcore.Core, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	var encoder zapcore.Encoder
	switch u.Query().Get("encoder") {
	case "json":
		encoder = zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	default:
		encoder = zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	}
	if u.Query().Get("filter") == "true" {
		encoder = &filter.FilterEncoder{Encoder: encoder}
	}

	lowPriority := zap.LevelEnablerFunc(func(level zapcore.Level) bool {
		return level < zapcore.ErrorLevel
	})
	highPriority := zap.LevelEnablerFunc(func(level zapcore.Level) bool {
		return level >= zapcore.ErrorLevel
	})
	lowWriter := zapcore.Lock(os.Stdout)
	highWriter := zapcore.Lock(os.Stderr)

	return zapcore.NewTee(
		zapcore.NewCore(encoder, lowWriter, lowPriority),
		zapcore.NewCore(encoder, highWriter, highPriority),
	), nil
}
