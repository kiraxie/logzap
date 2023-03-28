package logzap

import (
	"strings"

	"github.com/spf13/pflag"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	Level   zapcore.Level
	Modules ModulesLevel `yaml:",omitempty"`
}

func (t Config) RegisterFlagsWithPrefix(prefix string, f *pflag.FlagSet) {
	f.String(prefix+".level", zapcore.DebugLevel.String(), "Zap logger verbose level.")
}

type ModulesLevel map[string]zapcore.Level

func (t ModulesLevel) Get(s string) (l zapcore.Level) {
	if l, ok := t[strings.ToLower(s)]; ok {
		return l
	}

	return l
}
