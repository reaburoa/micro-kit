package server

// import (
// 	"github.com/go-kratos/kratos/v2/middleware/metadata"
// 	"time"

// 	prom "github.com/go-kratos/kratos/contrib/metrics/prometheus/v2"
// 	"github.com/go-kratos/kratos/v2/log"
// 	"github.com/go-kratos/kratos/v2/middleware/metrics"
// 	"github.com/go-kratos/kratos/v2/middleware/recovery"
// 	"github.com/go-kratos/kratos/v2/middleware/tracing"
// 	"github.com/go-kratos/kratos/v2/transport/grpc"
// 	"go.opentelemetry.io/otel/propagation"
// 	"tmc-gitlab.trasre.com/im/pkg/cloud/config"
// 	metrics2 "tmc-gitlab.trasre.com/im/pkg/cloud/metrics"
// 	"tmc-gitlab.trasre.com/im/pkg/middleware"
// 	"tmc-gitlab.trasre.com/im/pkg/utils/logs"
// )

// const (
// 	GrpcDefaultAddr    = "0.0.0.0:8081"
// 	GrpcDefaultTimeout = 3000 * time.Millisecond
// )

// type Grpc struct {
// 	Network string `json:"network,omitempty"`
// 	Addr    string `json:"addr,omitempty"`
// 	Timeout string `json:"timeout,omitempty"`
// }

// func GrpcServer(opts ...grpc.ServerOption) *grpc.Server {
// 	opts = append([]grpc.ServerOption{
// 		grpc.Middleware(
// 			recovery.Recovery(recovery.WithLogger(logs.GetAccessLog())),
// 			metrics.Server(
// 				metrics.WithSeconds(prom.NewHistogram(metrics2.ServerMetricSeconds)),
// 				metrics.WithRequests(prom.NewCounter(metrics2.ServerMetricRequests)),
// 			),
// 			tracing.Server(tracing.WithPropagator(propagation.NewCompositeTextMapPropagator(
// 				tracing.Metadata{}, propagation.TraceContext{}, propagation.Baggage{}))),
// 			middleware.ServerIErrorMiddleware(),
// 			middleware.AccessLogMiddleware(),
// 			metadata.Server(),
// 		),
// 		grpc.Logger(log.DefaultLogger),
// 		grpc.Address(GrpcDefaultAddr),
// 		grpc.Timeout(GrpcDefaultTimeout),
// 	}, opts...)
// 	var c Grpc
// 	if err := config.Scan("", "server.grpc", &c); err != nil {
// 		panic(err)
// 	}
// 	if c.Network != "" {
// 		opts = append(opts, grpc.Network(c.Network))
// 	}
// 	if c.Addr != "" {
// 		opts = append(opts, grpc.Address(c.Addr))
// 	}
// 	if c.Timeout != "" {
// 		duration, err := time.ParseDuration(c.Timeout)
// 		if err != nil {
// 			panic(err.Error())
// 		}
// 		opts = append(opts, grpc.Timeout(duration))
// 	}
// 	srv := grpc.NewServer(opts...)
// 	return srv
// }
