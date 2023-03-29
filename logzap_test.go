package logzap_test

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/kiraxie/logzap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
)

func testLog(log logzap.Logger, prefix string) {
	log.Trace(nil)
	//nolint: goerr113
	log.Trace(errors.New(prefix + "-trace"))
	log.Error(prefix + "-error")
	log.Errorf("%s-errorf-%d", prefix, 9527)
	log.Warn(prefix + "-warning")
	log.Info(prefix + "-info")
	log.Debug(prefix + "-debug")
}

func TestLogzap(t *testing.T) {
	t.Parallel()
	b := BufferZapCore{}
	require.NoError(t, logzap.Use(&b))
	require.NoError(t, logzap.Reload(logzap.Config{
		Level: zapcore.DebugLevel,
		Modules: logzap.ModulesLevel{
			"test1":  zapcore.DebugLevel,
			"test3":  zapcore.InfoLevel,
			"filter": zapcore.DebugLevel,
		},
	}))

	if log := logzap.Get("test1"); assert.NotNil(t, log) {
		testLog(log, "test1")
		assert.Contains(t, b.String(), "test1-trace")
		assert.Contains(t, b.String(), "test1-error")
		assert.Contains(t, b.String(), "test1-errorf-9527")
		assert.Contains(t, b.String(), "test1-warning")
		assert.Contains(t, b.String(), "test1-info")
		assert.Contains(t, b.String(), "test1-debug")
		assert.True(t, log.Enabled(zapcore.DebugLevel))
		assert.True(t, log.Enabled(zapcore.InfoLevel))
	}

	if log := logzap.Get("test3"); assert.NotNil(t, log) {
		log.Info("test5-info")
		log.Warn("test5-warn")
		log.Errorf("%v", fmt.Errorf("unknown"))

		assert.Contains(t, b.String(), "test5-info")
		assert.Contains(t, b.String(), "test5-warn")
		assert.Contains(t, b.String(), "unknown")
	}

	if log := logzap.Get("filter"); assert.NotNil(t, log) {
		log.Info(`Get "https://google.com?foo=boo&token=TOKEN-STRING": unexpected EOF`)
		assert.Contains(t, b.String(), `Get "https://google.com?foo=boo&token=[MASKED]": unexpected EOF`)
	}
}

type SyncBuffer struct {
	bytes.Buffer
}

func (t *SyncBuffer) Sync() error { return nil }

type BufferZapCore struct {
	Name string
	SyncBuffer
	Level zapcore.Level
}

func (t *BufferZapCore) With(fields []zapcore.Field) zapcore.Core { return t }

func (t *BufferZapCore) Enabled(level zapcore.Level) bool {
	return true
}

func (t *BufferZapCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if (ent.LoggerName == t.Name || t.Name == "") && t.Enabled(ent.Level) {
		return ce.AddCore(ent, t)
	}
	return ce
}

func (t *BufferZapCore) Write(ent zapcore.Entry, fields []zapcore.Field) (err error) {
	_, err = t.SyncBuffer.WriteString(logzap.FilterLogPattern(ent.Message) + "\n")
	return
}
