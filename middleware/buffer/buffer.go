package buffer

import (
	"bytes"
	"context"

	"github.com/kiraxie/logzap/middleware/console"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap/zapcore"
)

type syncBuffer struct {
	bytes.Buffer
}

func (t *syncBuffer) Sync() error { return nil }

type Buffer struct {
	Name string
	syncBuffer
	Level zapcore.Level
}

func (t *Buffer) With([]zapcore.Field) zapcore.Core { return t }

func (t *Buffer) Enabled(zapcore.Level) bool {
	return true
}

func (t *Buffer) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if (ent.LoggerName == t.Name || t.Name == "") && t.Enabled(ent.Level) {
		return ce.AddCore(ent, t)
	}

	return ce
}

func (t *Buffer) Write(ent zapcore.Entry, _ []zapcore.Field) (err error) {
	_, err = t.syncBuffer.WriteString(console.FilterLogPattern(ent.Message) + "\n")

	return
}

func New(
	_ context.Context,
	_ prometheus.Registerer,
	url string,
) (zapcore.Core, error) {
	return &Buffer{Name: url}, nil
}
