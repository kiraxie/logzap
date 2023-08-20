package filter

import (
	"regexp"

	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

func LogPattern(msg string) string {
	return reFilterToken.ReplaceAllString(msg, "${1}[MASKED]")
}

var reFilterToken = regexp.MustCompile(`([&?]token=)[0-9A-Za-z_-]+`)

type FilterEncoder struct {
	zapcore.Encoder
}

func (t *FilterEncoder) EncodeEntry(ent zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	ent.Message = LogPattern(ent.Message)

	return t.Encoder.EncodeEntry(ent, fields)
}
