package tracer

import (
	"fmt"
	"os"

	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
)

func TraceExporterWithStdout() (*stdouttrace.Exporter, error) {
	opts := []stdouttrace.Option{
		stdouttrace.WithWriter(os.Stdout),
	}
	exporter, err := stdouttrace.New(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout collector: %w", err)
	}

	// Set up a trace exporter
	return exporter, nil
}
