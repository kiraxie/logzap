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

func filterLogPattern(msg string) string {
	return reFilterToken.ReplaceAllString(msg, "${1}[MASKED]")
}

var reFilterToken = regexp.MustCompile(`([&?]token=)[0-9A-Za-z_-]+`)

type filterEncoder struct {
	zapcore.Encoder
}

func (t *filterEncoder) EncodeEntry(ent zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	ent.Message = filterLogPattern(ent.Message)
	return t.Encoder.EncodeEntry(ent, fields)
}

func New(
	_ context.Context,
	_ prometheus.Registerer,
	url string,
) (zapcore.Core, func(context.Context) error, error) {
	encoderConf := zap.NewDevelopmentEncoderConfig()
	enc := &filterEncoder{zapcore.NewConsoleEncoder(encoderConf)}
	sink, close, err := zap.Open(strings.Split(url, ",")...)
	if err != nil {
		return nil, nil, err
	}

	return zapcore.NewCore(enc, sink, zap.DebugLevel),
		func(context.Context) error {
			close()
			return nil
		}, nil
}
