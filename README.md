# logzap

A extension for uber/zap

## Purpose

The purpose of this package is make logger has sub module level verbose control.
In addition, lowest configuration and out of box.
I used Sugar logger of zap for all sub module, however it means lost some excellent performance to some degree.

## Usage

```go
logger := logzap.New(logzap.Config{
    Level: zapcore.WarnLevel,
    Modules: logzap.ModulesLevel{
        "foo":  zapcore.DebugLevel,
    },
})

fooLogger := logger.Get("foo") // debug level verbose
booLogger := logger.Get("boo") // warning level verbose
```
