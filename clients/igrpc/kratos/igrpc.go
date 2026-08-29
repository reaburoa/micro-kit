package kratos

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/go-kratos/kratos/v2/middleware/circuitbreaker"
	"github.com/go-kratos/kratos/v2/middleware/metadata"
	"github.com/go-kratos/kratos/v2/middleware/metrics"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/reaburoa/micro-kit/cloud/register/nacos"
	middleware "github.com/reaburoa/micro-kit/middleware/kratos"
	"go.opentelemetry.io/otel/propagation"
	goGrpc "google.golang.org/grpc"
)

func ConnGrpc(grpcServer string, options ...grpc.ClientOption) (*goGrpc.ClientConn, error) {
	return dialGrpc(grpcServer, options...)
}

func ConnGrpcByService(serviceName string, options ...grpc.ClientOption) (*goGrpc.ClientConn, error) {
	endpoint, err := resolveGrpcServiceEndpoint(serviceName)
	if err != nil {
		return nil, err
	}
	return dialGrpc(endpoint, options...)
}

func resolveGrpcServiceEndpoint(serviceName string) (string, error) {
	if serviceName == "" {
		return "", fmt.Errorf("grpc service name is empty")
	}
	client, err := nacos.NewServiceRegistryClient()
	if err != nil {
		return "", err
	}
	instance, err := client.SelectOneHealthyInstance(serviceName)
	if err != nil {
		return "", err
	}
	if instance == nil {
		return "", fmt.Errorf("grpc service %s resolve result is nil", serviceName)
	}
	return net.JoinHostPort(instance.Ip, strconv.FormatUint(instance.Port, 10)), nil
}

func dialGrpc(grpcServer string, options ...grpc.ClientOption) (*goGrpc.ClientConn, error) {
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
