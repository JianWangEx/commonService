# zap-extension

Some extensions for zap.

## Example

A custom encoder for zap. The encoder is used to print the trace ID to the log. Here's an basic example:

```golang
encoderConfig := zapext.NewProductionEncoderConfig()
// Configure time format.
encoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
    enc.AppendString(t.Format("2006-01-02T15:04:05.000-0700"))
}
// Configure the separator of console encoder.
encoderConfig.ConsoleSeparator = "|"

encoder := zapext.NewConsoleEncoder(encoderConfig)
core := zapcore.NewCore(encoder, os.Stderr, zapcore.DebugLevel)

logger := zap.New(core, zap.AddCaller())

// Add `@shopee_trace_id` to context.
logger = logger.With(zap.String("@shopee_trace_id", "XXXXX"))

logger.Info("test")
// 2021-06-29T15:31:29.927+0800|info|example/main.go:26|XXXXX|test
```
