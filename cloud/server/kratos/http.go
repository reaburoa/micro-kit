package kratos

import (
	"fmt"
	"time"

	"github.com/go-kratos/kratos/contrib/middleware/validate/v2"
	"github.com/go-kratos/kratos/v2/log"
	kmid "github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/metrics"
	"github.com/go-kratos/kratos/v2/middleware/ratelimit"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/reaburoa/micro-kit/cloud/config"
	"github.com/reaburoa/micro-kit/cloud/server"
	middleware "github.com/reaburoa/micro-kit/middleware/kratos"
	krtosLog "github.com/reaburoa/micro-kit/utils/log/kratos"
)

const (
	httpDefaultAddr    = ":8080"
	httpDefaultTimeout = 3000 * time.Millisecond
)

func NewHttp(middleware ...kmid.Middleware) *http.Server {
	return NewHttpWithName("http", middleware...)
}

func NewHttpWithName(srv string, middleware ...kmid.Middleware) *http.Server {
	var cfg server.Server
	if err := config.Get(fmt.Sprintf("server.%s", srv)).Scan(&cfg); err != nil {
		log.Fatalf("parse http %s config with error %s", srv, err.Error())
	}

	return newHttp(&cfg, middleware...)
}

func newHttp(conf *server.Server, kmiddleware ...kmid.Middleware) *http.Server {
	serverMiddleware := []kmid.Middleware{
		recovery.Recovery(),
		metrics.Server(
			metrics.WithRequests(middleware.MetricsRequests()),
			metrics.WithSeconds(middleware.MetricsSeconds()),
		),
		tracing.Server(),
		logging.Server(krtosLog.NewKratosLog()),
		validate.ProtoValidate(),
		ratelimit.Server(),
	}
	if len(kmiddleware) > 0 {
		serverMiddleware = append(serverMiddleware, kmiddleware...)
	}

	ops := []http.ServerOption{
		http.Middleware(serverMiddleware...),
		http.ResponseEncoder(middleware.CommonResponseFunc),
		http.ErrorEncoder(middleware.CommonErrorEncoder),
		http.Logger(krtosLog.NewKratosLog()),
		http.Address(httpDefaultAddr),
		http.Timeout(httpDefaultTimeout),
	}
	if conf.Network != "" {
		ops = append(ops, http.Network(conf.Network))
	}
	if conf.Port > 0 {
		ops = append(ops, http.Address(fmt.Sprintf(":%d", conf.Port)))
	}
	if conf.Timeout != "" {
		duration, err := time.ParseDuration(conf.Timeout)
		if err != nil {
			panic(err.Error())
		}
		ops = append(ops, http.Timeout(duration))
	}
	srv := http.NewServer(ops...)

	err := server.RunMetrics()
	if err != nil {
		log.Fatalf("running metrics failed with %s", err.Error())
	}

	err = server.RunPprof()
	if err != nil {
		log.Fatalf("running monitor failed with %s", err.Error())
	}

	return srv
}
