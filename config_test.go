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
		require.EqualValues(t, "level: error\n", b.String())
		require.NotContains(t, "modules:", b.String())
	})
	t.Run("unmarshal", func(t *testing.T) {
		t.Parallel()
		raw := `
level: "info"
modules:
  foo: "warn"
  boo: "debug"
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
	})
}
