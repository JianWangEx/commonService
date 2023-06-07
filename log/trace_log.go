// Package log @Author  wangjian    2023/6/2 12:20 AM
package log

import (
	"context"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

//func WithNewTraceLog(operationName string, ctx context.Context) (context.Context, tracing.Span) {
//	spanCtx := tracing.GetSpanContext(ctx)
//	if spanCtx == nil {
//		spanCtx = tracing.NewSpanContextGenerator("").NewSpanContext()
//	}
//	span, _ := tracing.GlobalTracer().NewSpan(operationName, spanCtx)
//	ctx = tracing.WithSpanContext(ctx, spanCtx)
//	newLogger := GetLogger().With(zap.String(grpczap.TraceKey, spanCtx.String()))
//	ctx = ctxzap.ToContext(ctx, newLogger)
//	return ctx, span
//}

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
	//if spanCtx := tracing.GetSpanContext(ctx); spanCtx != nil {
	//	return spanCtx.String()
	//}
	return ""
}
