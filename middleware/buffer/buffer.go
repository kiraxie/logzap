package buffer

import (
	"bytes"
	"context"

	"github.com/kiraxie/logzap/middleware/console"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap/zapcore"
)

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
	_, err = t.SyncBuffer.WriteString(console.FilterLogPattern(ent.Message) + "\n")
	return
}

func New(
	_ context.Context,
	_ prometheus.Registerer,
	url string,
) (zapcore.Core, error) {
	return &BufferZapCore{Name: url}, nil
}
