package tracer

import (
	"fmt"
	"io"
	"os"

	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
)

func TraceExporterWithStdout(stdoutEnable bool) (*stdouttrace.Exporter, error) {
	opts := []stdouttrace.Option{
		stdouttrace.WithWriter(io.Discard),
	}
	if stdoutEnable {
		opts = append(opts, stdouttrace.WithWriter(os.Stdout), stdouttrace.WithPrettyPrint())
	}
	exporter, err := stdouttrace.New(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout collector: %w", err)
	}

	// Set up a trace exporter
	return exporter, nil
}
