// Package log @Author  wangjian    2023/6/2 12:20 AM
package log

import (
	"context"
	"github.com/JianWangEx/commonService/log/zap-extension"
	"github.com/google/uuid"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

const (
	contextKeyForTraceId = "ti"
)

// WithNewTraceLog
//
//	@Description: 构建带有traceId的logger
//	@param ctx
//	@return context.Context
//
// TODO: 仅支持单个请求跟踪记录
func WithNewTraceLog(ctx context.Context) context.Context {
	traceId := uuid.New().String()
	ctx = context.WithValue(ctx, contextKeyForTraceId, traceId)
	newLogger := GetLogger().With(zap.String(zap_extension.TraceKey, traceId))
	ctx = ctxzap.ToContext(ctx, newLogger)
	return ctx
}

func GetTraceLogFromCtx(ctx context.Context) *zap.Logger {
	l := ctxzap.Extract(ctx)
	if l.Core().Enabled(zap.FatalLevel) {
		return l
	}
	return GetLogger()
}

func WithLogger(ctx context.Context, logger *zap.Logger) context.Context {
	return ctxzap.ToContext(ctx, logger)
}

func GetTraceIDFromCtx(ctx context.Context) string {
	traceId, ok := ctx.Value(contextKeyForTraceId).(string)
	if !ok {
		return ""
	}
	return traceId
}
