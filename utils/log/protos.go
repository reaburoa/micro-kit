package log

import (
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap/zapcore"
)

type SpanContext struct {
	TraceId      trace.TraceID
	SpanId       trace.SpanID
	ParentSpanId trace.SpanID
	Path         string
}

func (t *SpanContext) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if t.TraceId.IsValid() {
		enc.AddString("trace_id", t.TraceId.String())
	}
	if t.SpanId.IsValid() {
		enc.AddString("span_id", t.SpanId.String())
	}
	if t.ParentSpanId.IsValid() {
		enc.AddString("parent_span_id", t.ParentSpanId.String())
	}
	enc.AddString("path", t.Path)
	return nil
}

type Attributes struct {
	RequestId string
}

func (a *Attributes) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if a.RequestId != "" {
		enc.AddString("request_id", a.RequestId)
	}

	return nil
}

type Resource struct {
	CallMethod string
}

func (r *Resource) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if r.CallMethod != "" {
		enc.AddString("call_method", r.CallMethod)
	}

	return nil
}
