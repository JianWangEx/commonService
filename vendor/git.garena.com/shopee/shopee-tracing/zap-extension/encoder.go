package zaplib

import "go.uber.org/zap/zapcore"

// TraceKey is the key of trace field and used to access the log platform.
//
// Reference:
// https://confluence.shopee.io/display/LOG/%5BWIP%5DMake+your+log+structured
const TraceKey = "@shopee_trace_id"

// An EncoderConfig allows users to configure the concrete encoders supplied by
// zapcore.
//
// EncoderConfig warps the `zapcore.EncoderConfig` and carray the trace configration.
type EncoderConfig struct {
	TraceKey string `json:"traceKey" yaml:"traceKey"`
	zapcore.EncoderConfig
}

// NewProductionEncoderConfig returns an opinionated EncoderConfig for
// production environments.
func NewProductionEncoderConfig() EncoderConfig {
	return EncoderConfig{
		TraceKey: TraceKey,
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "ts",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.EpochTimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
	}
}
