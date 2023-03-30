package logzap_test

import (
	"bytes"
	"testing"

	"github.com/kiraxie/logzap"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v3"
)

func TestConfig(t *testing.T) {
	t.Parallel()
	t.Run("marshal", func(t *testing.T) {
		t.Parallel()
		config := &logzap.Config{
			Level: zapcore.ErrorLevel,
		}
		b := bytes.NewBuffer(nil)
		enc := yaml.NewEncoder(b)
		require.NoError(t, enc.Encode(config))
		require.Contains(t, b.String(), "level: error\n")
		require.NotContains(t, "modules:", b.String())
	})
	t.Run("unmarshal", func(t *testing.T) {
		t.Parallel()
		raw := `
level: "info"
modules:
  foo: "warn"
  boo: "debug"
middleware:
  console: "stdout"
  loki: "http://example.com:3100/loki/api/v1/push?label.instance=foo&label.job=boo"
`
		config := &logzap.Config{}
		dec := yaml.NewDecoder(bytes.NewReader([]byte(raw)))
		require.NoError(t, dec.Decode(config))
		require.EqualValues(t, zapcore.InfoLevel, config.Level)

		require.Len(t, config.Modules, 2)
		require.Contains(t, config.Modules, "foo")
		require.EqualValues(t, zapcore.WarnLevel, config.Modules["foo"])
		require.Contains(t, config.Modules, "boo")
		require.EqualValues(t, zapcore.DebugLevel, config.Modules["boo"])

		require.Len(t, config.Middleware, 2)
		require.Contains(t, config.Middleware, "console")
		require.EqualValues(t, "stdout", config.Middleware["console"])
		require.Contains(t, config.Middleware, "loki")
		require.EqualValues(t, "http://example.com:3100/loki/api/v1/push?label.instance=foo&label.job=boo", config.Middleware["loki"])
	})
}
