package logzap

import (
	"regexp"

	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

type filterEncoder struct {
	zapcore.Encoder
}

func (t *filterEncoder) EncodeEntry(ent zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	ent.Message = FilterLogPattern(ent.Message)

	return t.Encoder.EncodeEntry(ent, fields)
}

func FilterLogPattern(msg string) string {
	msg = reFilterToken.ReplaceAllString(msg, "${1}[MASKED]")

	return msg
}

var reFilterToken = regexp.MustCompile(`([&?]token=)[0-9A-Za-z_-]+`)
