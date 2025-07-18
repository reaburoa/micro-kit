package kratos

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/middleware/circuitbreaker"
	"github.com/go-kratos/kratos/v2/middleware/metadata"
	"github.com/go-kratos/kratos/v2/middleware/metrics"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	middleware "github.com/reaburoa/micro-kit/middleware/kratos"
	"go.opentelemetry.io/otel/propagation"
	goGrpc "google.golang.org/grpc"
)

func ConnGrpc(grpcServer string, options ...grpc.ClientOption) (*goGrpc.ClientConn, error) {
	options = append([]grpc.ClientOption{
		grpc.WithTimeout(3 * time.Second),
		grpc.WithEndpoint(grpcServer),
		grpc.WithMiddleware(
			recovery.Recovery(),
			middleware.RequestLogMiddleware(),
			middleware.ClientErrorMiddleware(),
			metadata.Client(),
			tracing.Client(tracing.WithPropagator(propagation.NewCompositeTextMapPropagator(
				tracing.Metadata{}, propagation.TraceContext{}, propagation.Baggage{}))),
			metrics.Client(
				metrics.WithRequests(middleware.MetricsRequests()),
				metrics.WithSeconds(middleware.MetricsSeconds()),
			),
			circuitbreaker.Client(),
		),
	}, options...)
	return grpc.DialInsecure(
		context.Background(),
		options...,
	)
}
