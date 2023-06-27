package console

import (
	"context"
	"regexp"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

func FilterLogPattern(msg string) string {
	return reFilterToken.ReplaceAllString(msg, "${1}[MASKED]")
}

var reFilterToken = regexp.MustCompile(`([&?]token=)[0-9A-Za-z_-]+`)

type filterEncoder struct {
	zapcore.Encoder
}

func (t *filterEncoder) EncodeEntry(ent zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	ent.Message = FilterLogPattern(ent.Message)

	return t.Encoder.EncodeEntry(ent, fields)
}

func New(
	_ context.Context,
	_ prometheus.Registerer,
	url string,
) (zapcore.Core, error) {
	if url == "" {
		return zapcore.NewNopCore(), nil
	}
	encoderConf := zap.NewDevelopmentEncoderConfig()
	enc := &filterEncoder{zapcore.NewConsoleEncoder(encoderConf)}
	sink, _, err := zap.Open(strings.Split(url, ",")...)
	if err != nil {
		return nil, err
	}

	return zapcore.NewCore(enc, sink, zap.DebugLevel), nil
}
