package logzap

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	ErrInvalidModuleLevel = fmt.Errorf("invalid module level")
	ErrUnsupportedFields  = fmt.Errorf("unsupported fields")
)

type Config struct {
	Level   zapcore.Level `yaml:"level"`
	Modules ModulesLevel  `yaml:"modules,omitempty"`
	Cores   Cores         `yaml:"cores,omitempty"`
}

func (t Config) RegisterFlagsWithPrefix(prefix string, f *pflag.FlagSet) {
	f.String(prefix+".level", zapcore.DebugLevel.String(), "Global logger verbose level.")
}

type ModulesLevel map[string]zapcore.Level

func (t ModulesLevel) Get(s string) (l zapcore.Level) {
	if l, ok := t[strings.ToLower(s)]; ok {
		return l
	}

	return l
}

func (t ModulesLevel) build(log *zap.Logger) map[string]*logger {
	m := map[string]*logger{}
	for k, lv := range t {
		name := strings.ToLower(k)
		m[name] = newLogger(log.Named(name), lv)
	}

	return m
}

var MapStructureLevelDecodeHook = []mapstructure.DecodeHookFunc{
	levelDecodeHookFunc,
	mapStringDecodeHookFunc,
}

func mapStringDecodeHookFunc(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if t != typeModuleLevel || f.Kind() != reflect.Map ||
		f.Key().Kind() != reflect.String || f.Elem().Kind() != reflect.Interface {
		return data, nil
	}
	if v, err := parseModulesLevel(data); err == nil {
		return v, nil
	}
	if v, err := parseMiddleware(data); err == nil {
		return v, nil
	}

	return nil, fmt.Errorf("%w: %#v", ErrUnsupportedFields, data)
}

func levelDecodeHookFunc(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if t != typeLevel || f.Kind() != reflect.String {
		return data, nil
	}
	var level zapcore.Level
	err := level.UnmarshalText([]byte(data.(string)))

	return level, err
}

func parseModulesLevel(data interface{}) (result ModulesLevel, err error) {
	source, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("%w: %#v", ErrInvalidModuleLevel, data)
	}
	result = ModulesLevel{}
	for key, rawLevel := range source {
		module := strings.ToLower(key)
		switch lv := rawLevel.(type) {
		case string:
			var level zapcore.Level
			err := level.UnmarshalText([]byte(lv))
			if err != nil {
				return nil, fmt.Errorf("%w: %#v", err, lv)
			}
			result[module] = level
		case int:
			result[module] = zapcore.Level(lv)
		case map[string]interface{}:
			child, err := parseModulesLevel(lv)
			if err != nil {
				return nil, err
			}
			for name, lv := range child {
				result[module+"."+name] = lv
			}
		default:
			return nil, fmt.Errorf("%w: %#v", ErrInvalidModuleLevel, data)
		}
	}

	return result, nil
}

func parseMiddleware(data interface{}) (result Cores, err error) {
	source, ok := data.(map[string]string)
	if !ok {
		return nil, fmt.Errorf("%w: %#v", ErrInvalidModuleLevel, data)
	}
	result = Cores{}
	for key, url := range source {
		result[strings.ToLower(key)] = url
	}

	return result, nil
}

var (
	typeLevel       = reflect.TypeOf(zap.DebugLevel)
	typeModuleLevel = reflect.TypeOf(ModulesLevel{})
)
