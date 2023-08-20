package logzap_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/kiraxie/logzap"
	"github.com/kiraxie/logzap/core/buffer"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
)

var (
	ErrUnknown = errors.New("unknown")
	ErrTest    = errors.New("test")
)

func BenchmarkLogzap(b *testing.B) {
	require := require.New(b)
	require.NotNil(require)

	l := logzap.New(context.Background(), prometheus.DefaultRegisterer, logzap.Config{Level: zapcore.DebugLevel})
	core, err := buffer.New(context.Background(), prometheus.DefaultRegisterer, "")
	require.NoError(err)
	log := l.Get("benchmark")
	l.Use(core)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		log.Info(".")
	}
	b.StopTimer()

	require.IsType(&buffer.Buffer{}, core)
	buff := core.(*buffer.Buffer)
	require.NotZero(buff.Len())
}

func testLog(log logzap.Logger, prefix string) {
	log.Trace(nil)
	log.Trace(fmt.Errorf("%w %s", ErrTest, prefix+"-trace"))
	log.Error(prefix + "-error")
	log.Errorf("%s-errorf-%d", prefix, 9527)
	log.Warn(prefix + "-warning")
	log.Info(prefix + "-info")
	log.Debug(prefix + "-debug")
}

func TestLogzap(t *testing.T) {
	t.Parallel()
	logger := logzap.New(
		context.Background(), prometheus.DefaultRegisterer,
		logzap.Config{
			Level: zapcore.WarnLevel,
			Modules: logzap.ModulesLevel{
				"test1":  zapcore.DebugLevel,
				"test3":  zapcore.InfoLevel,
				"filter": zapcore.DebugLevel,
			},
		},
	)
	core, err := buffer.New(context.Background(), prometheus.DefaultRegisterer, "")
	require.NoError(t, err)
	require.IsType(t, &buffer.Buffer{}, core)
	b := core.(*buffer.Buffer)
	logger.Use(core)

	if log := logger.Get("test1"); assert.NotNil(t, log) {
		testLog(log, "test1")
		assert.Contains(t, b.String(), "test1-trace")
		assert.Contains(t, b.String(), "test1-error")
		assert.Contains(t, b.String(), "test1-errorf-9527")
		assert.Contains(t, b.String(), "test1-warning")
		assert.Contains(t, b.String(), "test1-info")
		assert.Contains(t, b.String(), "test1-debug")
	}

	if log := logger.Get("test3"); assert.NotNil(t, log) {
		log.Info("test5-info")
		log.Warn("test5-warn")
		log.Errorf("%w", ErrUnknown)

		assert.Contains(t, b.String(), "test5-info")
		assert.Contains(t, b.String(), "test5-warn")
		assert.Contains(t, b.String(), "unknown")
	}

	if log := logger.Get("filter"); assert.NotNil(t, log) {
		log.Info(`Get "https://google.com?foo=boo&token=TOKEN-STRING": unexpected EOF`)
		assert.Contains(t, b.String(), `Get "https://google.com?foo=boo&token=[MASKED]": unexpected EOF`)
	}
}
