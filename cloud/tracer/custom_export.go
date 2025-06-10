package tracer

import (
	"context"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type ITraceExporter struct {
}

func (i *ITraceExporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	return nil
}

func (i *ITraceExporter) Shutdown(ctx context.Context) error {
	return nil
}
