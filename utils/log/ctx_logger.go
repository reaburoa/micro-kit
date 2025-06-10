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

func CtxDebug(ctx context.Context, a ...interface{}) {
	msg := getMessage("", a)
	zap.L().Debug(msg, buildCtxField(ctx)...)
}

func CtxDebugf(ctx context.Context, format string, a ...interface{}) {
	msg := getMessage(format, a)
	zap.L().Debug(msg, buildCtxField(ctx)...)
}

func CtxInfo(ctx context.Context, a ...interface{}) {
	msg := getMessage("", a)
	zap.L().Info(msg, buildCtxField(ctx)...)
}

// // CtxInfof uses fmt.Sprintf to log a templated message.
func CtxInfof(ctx context.Context, format string, a ...interface{}) {
	msg := getMessage(format, a)
	zap.L().Info(msg, buildCtxField(ctx)...)
}

func CtxWarn(ctx context.Context, a ...interface{}) {
	msg := getMessage("", a)
	zap.L().Warn(msg, buildCtxField(ctx)...)
}

func CtxWarnf(ctx context.Context, format string, a ...interface{}) {
	msg := getMessage(format, a)
	zap.L().Warn(msg, buildCtxField(ctx)...)
}

func CtxError(ctx context.Context, a ...interface{}) {
	msg := getMessage("", a)
	zap.L().Error(msg, buildCtxField(ctx)...)
}

func CtxErrorf(ctx context.Context, format string, a ...interface{}) {
	msg := getMessage(format, a)
	zap.L().Error(msg, buildCtxField(ctx)...)
}

func CtxFatal(ctx context.Context, a ...interface{}) {
	msg := getMessage("", a)
	zap.L().Fatal(msg, buildCtxField(ctx)...)
}

func CtxFatalf(ctx context.Context, format string, a ...interface{}) {
	msg := getMessage(format, a)
	zap.L().Fatal(msg, buildCtxField(ctx)...)
}
