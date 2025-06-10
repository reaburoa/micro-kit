package tracer

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/welltop-cn/common/cloud/config"
	"github.com/welltop-cn/common/protos"
	"github.com/welltop-cn/common/utils/env"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	domain = "listenbook"
)

var (
	TraceProvider trace.Tracer
)

func InitOtelTracer() (func(context.Context) error, error) {
	var (
		err      error
		exporter sdktrace.SpanExporter
	)
	cfg, cfgErr := GetExporterConfig()
	if cfgErr != nil {
		//log.L().Errorf("get trace exporter config: %#v", cfgErr)
	}
	if !env.IsDebug() {
		exporter, err = TraceExporterWithGrpc(context.Background(), cfg.Target)
	} else {
		exporter, err = TraceExporterWithStdout()
	}
	if err != nil {
		//log.L().Errorf("failed to create trace exporter: %#v", err)
		return nil, err
	}
	return InitProvider(exporter, cfg.Sample)
}

// Initializes an OTLP exporter, and configures the corresponding trace providers.
func InitProvider(exporter sdktrace.SpanExporter, sampleRate float64) (func(context.Context) error, error) {
	traceService := os.Getenv("JAEGER_SERVICE_NAME")
	if traceService == "" {
		traceService = fmt.Sprintf("%s-%s", domain, env.ServiceName())
	}
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			// the service name used to display traces in backends
			semconv.ServiceName(traceService),
			attribute.String("env", string(env.GetRuntimeEnv())),
			attribute.String("region", string(env.GetRuntimeRegion())),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	opts := []sdktrace.BatchSpanProcessorOption{
		sdktrace.WithMaxQueueSize(50000),
		sdktrace.WithBatchTimeout(time.Second * 3),
	}
	sample := sdktrace.ParentBased(sdktrace.AlwaysSample())
	if env.IsRelease() && sampleRate > 0 {
		sample = sdktrace.ParentBased(sdktrace.TraceIDRatioBased(sampleRate))
	}
	// Register the trace exporter with a TracerProvider, using a batch
	// span processor to aggregate spans before export.
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sample),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(sdktrace.NewBatchSpanProcessor(exporter, opts...)),
	)
	otel.SetTracerProvider(tracerProvider)

	// set global propagator to tracecontext (the default is no-op).
	otel.SetTextMapPropagator(propagation.TraceContext{})

	TraceProvider = GetTracer()
	// Shutdown will flush any remaining spans and shut down the exporter.
	return tracerProvider.Shutdown, nil
}

func GetTextMapPropagator() propagation.TextMapPropagator {
	return otel.GetTextMapPropagator()
}

func GetTracerProvider() trace.TracerProvider {
	return otel.GetTracerProvider()
}

func GetTracer() trace.Tracer {
	return otel.Tracer(env.ServiceName())
}

func GetExporterConfig() (*protos.TracerExporter, error) {
	var cfg *protos.TracerExporter
	err := config.Get("tracer").Scan(&cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
