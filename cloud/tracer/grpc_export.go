package tracer

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func TraceExporterWithGrpc(ctx context.Context, target string, ops ...grpc.DialOption) (*otlptrace.Exporter, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	defaultOps := []grpc.DialOption{
		grpc.WithTransportCredentials(credentials.NewTLS(nil)), // Note the use of insecure transport here. TLS is recommended in production.
		grpc.WithBlock(),
	}
	if len(ops) > 0 {
		defaultOps = append(defaultOps, ops...)
	}
	// If the OpenTelemetry Collector is running on a local cluster (minikube or
	// microk8s), it should be accessible through the NodePort service at the
	// `localhost:30080` endpoint. Otherwise, replace `localhost` with the
	// endpoint of your cluster. If you run the app inside k8s, then you can
	// probably connect directly to the service through dns.
	conn, err := grpc.DialContext(ctx, target, defaultOps...)
	//log.L().Info("dial tracing collector ==> ", conn, err)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection to collector: %w", err)
	}
	// Set up a trace exporter
	traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	//log.L().Info("traceExporter collector ==>", traceExporter, err)
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	return traceExporter, nil
}
