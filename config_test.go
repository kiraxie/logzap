package logzap_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/kiraxie/logzap"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
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
cores:
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

		require.Len(t, config.Cores, 1)
		require.Contains(t, config.Cores, "loki")
		require.EqualValues(t, "http://example.com:3100/loki/api/v1/push?label.instance=foo&label.job=boo", config.Cores["loki"])
	})
}

func TestViper(t *testing.T) {
	t.Parallel()
	v := viper.New()
	v.SetConfigType("yaml")
	err := v.ReadConfig(strings.NewReader(strings.ReplaceAll(strings.TrimSpace(`
level: WARN
modules:
	test1:              debug
	test2:              2
	caseSensitiveTest1: debug
	caseSensitiveTest2: 2
	parent:
		child:          error
cores:
	console: stdout
	loki: example.com
`), "\t", "  ")))
	require.NoError(t, err)

	config := &logzap.Config{}
	require.NoError(t, v.Unmarshal(config, decoderConfigOption))

	require.Len(t, config.Modules, 5)
	require.Contains(t, config.Modules, "test1")
	require.Contains(t, config.Modules, "test2")
	require.Contains(t, config.Modules, "casesensitivetest1")
	require.Contains(t, config.Modules, "casesensitivetest2")
	require.Contains(t, config.Modules, "parent.child")

	require.Len(t, config.Cores, 2)
	require.Contains(t, config.Cores, "console")
	require.Contains(t, config.Cores, "loki")
}

var decoderConfigOption = func(hookFuncs ...mapstructure.DecodeHookFunc) func(cfg *mapstructure.DecoderConfig) {
	return func(cfg *mapstructure.DecoderConfig) {
		if cfg.DecodeHook == nil {
			cfg.DecodeHook = mapstructure.ComposeDecodeHookFunc(hookFuncs...)
		} else {
			cfg.DecodeHook = mapstructure.ComposeDecodeHookFunc(append(
				[]mapstructure.DecodeHookFunc{cfg.DecodeHook}, hookFuncs...)...)
		}
	}
}(logzap.MapStructureLevelDecodeHook...)
