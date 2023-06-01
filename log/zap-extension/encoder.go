// Package zap_extension @Author  wangjian    2023/6/1 5:25 PM
package zap_extension

import "go.uber.org/zap/zapcore"

const TraceKey = "@stupid_trace_id"

// EncoderConfig 在zapcore.EncoderConfig 基础上新增了trace配置
type EncoderConfig struct {
	TraceKey string `json:"traceKey" yaml:"traceKey"`
	zapcore.EncoderConfig
}

// NewProductionEncoderConfig 构造了一个固定的EncoderConfig
func NewProductionEncoderConfig() EncoderConfig {
	return EncoderConfig{
		TraceKey: TraceKey,
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "ts",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "message",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.CapitalLevelEncoder,
			EncodeTime:     zapcore.EpochTimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
	}
}
