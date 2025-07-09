package kratos

import (
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/middleware/metadata"
	"github.com/go-kratos/kratos/v2/middleware/metrics"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/reaburoa/micro-kit/cloud/config"
	"github.com/reaburoa/micro-kit/cloud/server"
	middleware "github.com/reaburoa/micro-kit/middleware/kratos"
	"github.com/reaburoa/micro-kit/utils/log"
	krtosLog "github.com/reaburoa/micro-kit/utils/log/kratos"
)

const (
	grpcDefaultAddr    = ":8081"
	grpcDefaultTimeout = 3000 * time.Millisecond
)

func NewGrpc(nacosOpts ...grpc.ServerOption) *grpc.Server {
	return NewGrpcWithName("grpc", nacosOpts...)
}

// NewGrpcWithName 启动指定的grpc服务
func NewGrpcWithName(grpcSrv string, nacosOpts ...grpc.ServerOption) *grpc.Server {
	var cfg *server.Server
	if err := config.Get(fmt.Sprintf("server.%s", grpcSrv)).Scan(&cfg); err != nil {
		log.Error(err)
	}
	return newGrpc(cfg, nacosOpts...)
}

func newGrpc(conf *server.Server, opts ...grpc.ServerOption) *grpc.Server {
	opts = append([]grpc.ServerOption{
		grpc.Middleware(
			recovery.Recovery(),
			metrics.Server(
				metrics.WithRequests(middleware.MetricsRequests()),
				metrics.WithSeconds(middleware.MetricsSeconds()),
			),
			tracing.Server(),
			middleware.ServerErrorMiddleware(),
			middleware.AccessLogMiddleware(),
			metadata.Server(),
		),
		grpc.Logger(krtosLog.NewKratosLog()),
		grpc.Address(grpcDefaultAddr),
		grpc.Timeout(grpcDefaultTimeout),
	}, opts...)
	if conf.Network != "" {
		opts = append(opts, grpc.Network(conf.Network))
	}
	if conf.Port > 0 {
		opts = append(opts, grpc.Address(fmt.Sprintf(":%d", conf.Port)))
	}
	if conf.Timeout != "" {
		duration, err := time.ParseDuration(conf.Timeout)
		if err != nil {
			panic(err.Error())
		}
		opts = append(opts, grpc.Timeout(duration))
	}
	srv := grpc.NewServer(opts...)
	return srv
}
