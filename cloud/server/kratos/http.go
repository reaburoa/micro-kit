package server

import (
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	kmid "github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/reaburoa/micro-kit/cloud/config"
	"github.com/reaburoa/micro-kit/cloud/server"
	middleware "github.com/reaburoa/micro-kit/middleware/kratos"
)

const (
	HTTPDefaultAddr    = "0.0.0.0:8080"
	HTTPDefaultTimeout = 3000 * time.Millisecond
)

var DefaultMiddlewares = []kmid.Middleware{
	// recovery.Recovery(recovery.WithLogger(logs.GetAccessLog())),
	// metrics.Server(
	// 	metrics.WithSeconds(prom.NewHistogram(metrics2.ServerMetricSeconds)),
	// 	metrics.WithRequests(prom.NewCounter(metrics2.ServerMetricRequests)),
	// ),
	// tracing.Server(
	// 	tracing.WithPropagator(propagation.NewCompositeTextMapPropagator(
	// 		tracing.Metadata{}, propagation.TraceContext{}, propagation.Baggage{})),
	// ),
	// middleware.ServerIErrorMiddleware(),
	// metadata.Server(),
	// middleware.AccessLogMiddleware(),
	// validate.Validator(),
	// ratelimit.Server(),
	// middleware.CommonHeaderMiddleware(),
}

func AddMiddleware(middlewares ...kmid.Middleware) {
	DefaultMiddlewares = append(DefaultMiddlewares, middlewares...)
}

func NewHttp(ops ...http.ServerOption) *http.Server {
	return NewHttpWithName("http", ops...)
}

func NewHttpWithName(srv string, ops ...http.ServerOption) *http.Server {
	var cfg server.Server
	if err := config.Get(fmt.Sprintf("server.%s", srv)).Scan(&cfg); err != nil {
		log.Fatalf("parse http %s config with error %s", srv, err.Error())
	}

	return newHttp(&cfg, ops...)
}

func newHttp(conf *server.Server, ops ...http.ServerOption) *http.Server {
	ops = append([]http.ServerOption{
		http.Middleware(DefaultMiddlewares...),
		http.ResponseEncoder(middleware.CommonResponseFunc),
		http.ErrorEncoder(middleware.CommonErrorEncoder),
		http.Logger(log.DefaultLogger),
		http.Address(HTTPDefaultAddr),
		http.Timeout(HTTPDefaultTimeout),
	}, ops...)

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

	return srv
}
