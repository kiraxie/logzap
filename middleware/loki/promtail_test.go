package loki_test

import (
	"context"
	"testing"
	"time"

	"github.com/kiraxie/logzap/middleware/loki"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestPromtail(t *testing.T) {
	t.Parallel()
	core, shutdown, err := loki.New(
		context.Background(),
		prometheus.DefaultRegisterer,
		"http://example.com:3100/loki/api/v1/push?dryRun=true&label.instance=foo&label.job=boo",
	)
	require.NoError(t, err)
	defer shutdown(context.Background())
	logger := zap.New(core)
	logger.Error("abc")
	logger.Error("123")
	logger.Named("foo").Info("123")
	time.Sleep(500 * time.Millisecond)
}
