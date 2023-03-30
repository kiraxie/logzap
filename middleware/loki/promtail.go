package loki

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	retry "github.com/cenkalti/backoff/v4"
	"github.com/go-kit/log"
	"github.com/grafana/dskit/backoff"
	"github.com/grafana/dskit/flagext"
	"github.com/grafana/loki/clients/pkg/promtail/api"
	promtail "github.com/grafana/loki/clients/pkg/promtail/client"
	"github.com/grafana/loki/pkg/push"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/model"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func defaultPromtailConfig() promtail.Config {
	return promtail.Config{
		BackoffConfig: backoff.Config{
			MaxBackoff: promtail.MaxBackoff,
			MaxRetries: promtail.MaxRetries,
			MinBackoff: promtail.MinBackoff,
		},
		BatchSize: promtail.BatchSize,
		BatchWait: promtail.BatchWait,
		Timeout:   promtail.Timeout,
	}
}

type Client struct {
	ctx context.Context
	promtail.Client
	log.Logger
	encoder zapcore.Encoder
	label   model.LabelSet
}

func New(
	ctx context.Context,
	registry prometheus.Registerer,
	rawURL string,
) (zapcore.Core, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	t := &Client{
		ctx:   ctx,
		label: model.LabelSet{},
	}
	name := ""
	encoding := ""
	dryRun := false
	for k, v := range u.Query() {
		switch {
		case k == "name" && len(v) != 0:
			name = v[0]
		case k == "encoding" && len(v) != 0:
			encoding = v[0]
		case k == "dryRun" && (len(v) == 0 || v[0] == "true"):
			dryRun = true
		case strings.HasPrefix(k, "label.") && len(v) != 0:
			t.label[model.LabelName(strings.TrimPrefix(k, "label."))] = model.LabelValue(v[0])
		default:
		}
	}
	u.RawQuery = ""
	promtailConf := defaultPromtailConfig()
	promtailConf.URL = flagext.URLValue{URL: u}
	promtailConf.Name = name

	if t.Client, err = newPromtailClient(ctx, registry, promtailConf, t.Logger, dryRun); err != nil {
		return nil, err
	}

	t.encoder = newZapEncoder(encoding)

	return t, nil
}

func (t *Client) With(fields []zapcore.Field) zapcore.Core { return t }

func (t *Client) Enabled(zapcore.Level) bool {
	// allow all incoming log
	return true
}

func (t *Client) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	return ce.AddCore(ent, t)
}

func (t *Client) Write(ent zapcore.Entry, fields []zapcore.Field) error {
	select {
	case <-t.ctx.Done():
		return nil
	default:
	}

	msg, err := t.encoder.EncodeEntry(ent, fields)
	if err != nil {
		return err
	}
	e := api.Entry{
		Labels: t.label.Clone(),
		Entry:  push.Entry{Timestamp: ent.Time, Line: msg.String()},
	}

	return retry.Retry(func() error {
		select {
		case <-t.ctx.Done():
			return nil
		case t.Chan() <- e:
			return nil
		default:
		}
		return fmt.Errorf("channel full")
	}, retry.WithContext(&retry.ExponentialBackOff{
		InitialInterval:     100 * time.Millisecond,
		RandomizationFactor: 0,
		Multiplier:          1.5,
		MaxInterval:         2 * time.Second,
		MaxElapsedTime:      5 * time.Second,
		Stop:                -1,
		Clock:               retry.SystemClock,
	}, t.ctx))
}

func (t *Client) Sync() error {
	t.Client.Stop()
	return nil
}

func newPromtailClient(
	ctx context.Context,
	registry prometheus.Registerer,
	config promtail.Config,
	logger log.Logger,
	dryRun bool,
) (client promtail.Client, err error) {
	metrics := promtail.NewMetrics(registry)
	if dryRun {
		client, err = promtail.NewLogger(metrics, logger, config)
	} else {
		client, err = promtail.New(metrics, config, 0, 0, dryRun, logger)
	}
	if err != nil {
		return nil, err
	}

	return
}

func newZapEncoder(encoding string) (enc zapcore.Encoder) {
	zEncConf := zap.NewProductionEncoderConfig()
	zEncConf.TimeKey = ""
	zEncConf.EncodeTime = zapcore.ISO8601TimeEncoder
	switch encoding {
	case "json":
		enc = zapcore.NewJSONEncoder(zEncConf)
	case "console":
		enc = zapcore.NewConsoleEncoder(zEncConf)
	default:
		enc = zapcore.NewJSONEncoder(zEncConf)
	}
	return
}
