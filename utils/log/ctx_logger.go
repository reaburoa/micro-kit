package log

import (
	"context"
	"fmt"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

func getMessage(template string, fmtArgs []interface{}) string {
	if len(fmtArgs) == 0 {
		return template
	}

	if template != "" {
		return fmt.Sprintf(template, fmtArgs...)
	}

	if len(fmtArgs) == 1 {
		if str, ok := fmtArgs[0].(string); ok {
			return str
		}
	}
	return fmt.Sprint(fmtArgs...)
}

func buildSpanField(ctx context.Context) zap.Field {
	spanCtx := &SpanContext{}

	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return zap.Object(CtxField, spanCtx)
	}
	if rspan, ok := span.(sdktrace.ReadOnlySpan); ok {
		traceId := rspan.SpanContext().TraceID()
		spanId := rspan.SpanContext().SpanID()
		var pid trace.SpanID
		if rspan.Parent().IsValid() {
			pid = rspan.Parent().SpanID()
		}
		spanCtx = &SpanContext{
			TraceId:      traceId,
			SpanId:       spanId,
			ParentSpanId: pid,
			Path:         rspan.Name(),
		}
		return zap.Object(CtxField, spanCtx)
	}

	return zap.Object(CtxField, spanCtx)
}

func buildAttrField(ctx context.Context) zap.Field {
	attr := &Attributes{}
	if reqId, ok := ctx.Value(AttrRequestId).(string); ok {
		attr.RequestId = reqId
	}
	return zap.Object(AttributesField, attr)
}

func buildCtxField(ctx context.Context) []zap.Field {
	zapFields := make([]zap.Field, 0, 3)
	if ctx != nil {
		zapFields = append(zapFields, buildSpanField(ctx), buildAttrField(ctx))
	}

	return zapFields
}

func ctxLog(ctx context.Context, level Level, template string, args ...interface{}) {
	zapLevel := Level2ZapLevle(level)
	msg := getMessage(template, args)
	kvs := make([]interface{}, 0, 20)
	for _, value := range buildCtxField(ctx) {
		kvs = append(kvs, value.Key, value.Interface)
	}
	zap.S().Logw(zapLevel, msg, kvs...)
}

func ctxLogW(ctx context.Context, level Level, msg string, keysAndValues ...interface{}) {
	zapLevel := Level2ZapLevle(level)
	kvs := make([]interface{}, 0, 20)
	kvs = append(kvs, keysAndValues...)
	for _, value := range buildCtxField(ctx) {
		kvs = append(kvs, value.Key, value.Interface)
	}
	zap.S().Logw(zapLevel, msg, kvs...)
}

func CtxDebug(ctx context.Context, a ...interface{}) {
	ctxLog(ctx, DebugLevel, "", a...)
}

func CtxDebugf(ctx context.Context, format string, a ...interface{}) {
	ctxLog(ctx, DebugLevel, format, a...)
}

func CtxDebugw(ctx context.Context, msg string, a ...interface{}) {
	ctxLogW(ctx, DebugLevel, msg, a...)
}

func CtxInfo(ctx context.Context, a ...interface{}) {
	ctxLog(ctx, InfoLevel, "", a...)
}

func CtxInfof(ctx context.Context, format string, a ...interface{}) {
	ctxLog(ctx, InfoLevel, format, a...)
}

func CtxInfow(ctx context.Context, msg string, a ...interface{}) {
	ctxLogW(ctx, InfoLevel, msg, a...)
}

func CtxWarn(ctx context.Context, a ...interface{}) {
	ctxLog(ctx, WarnLevel, "", a...)
}

func CtxWarnf(ctx context.Context, format string, a ...interface{}) {
	ctxLog(ctx, WarnLevel, format, a...)
}

func CtxWarnw(ctx context.Context, msg string, a ...interface{}) {
	ctxLogW(ctx, WarnLevel, msg, a...)
}

func CtxError(ctx context.Context, a ...interface{}) {
	ctxLog(ctx, ErrorLevel, "", a...)
}

func CtxErrorf(ctx context.Context, format string, a ...interface{}) {
	ctxLog(ctx, ErrorLevel, format, a...)
}

func CtxErrorw(ctx context.Context, msg string, a ...interface{}) {
	ctxLogW(ctx, ErrorLevel, msg, a...)
}

func CtxFatal(ctx context.Context, a ...interface{}) {
	ctxLog(ctx, FatalLevel, "", a...)
}

func CtxFatalf(ctx context.Context, format string, a ...interface{}) {
	ctxLog(ctx, FatalLevel, format, a...)
}

func CtxFatalw(ctx context.Context, msg string, a ...interface{}) {
	ctxLogW(ctx, FatalLevel, msg, a...)
}
