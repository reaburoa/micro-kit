package server

// import (
// 	"context"
// 	"errors"
// 	netHttp "net/http"
// 	"strings"
// 	"time"

// 	prom "github.com/go-kratos/kratos/contrib/metrics/prometheus/v2"
// 	"github.com/go-kratos/kratos/v2/log"
// 	kmid "github.com/go-kratos/kratos/v2/middleware"
// 	kratosMiddleware "github.com/go-kratos/kratos/v2/middleware"
// 	"github.com/go-kratos/kratos/v2/middleware/metadata"
// 	"github.com/go-kratos/kratos/v2/middleware/metrics"
// 	"github.com/go-kratos/kratos/v2/middleware/ratelimit"
// 	"github.com/go-kratos/kratos/v2/middleware/recovery"
// 	"github.com/go-kratos/kratos/v2/middleware/tracing"
// 	"github.com/go-kratos/kratos/v2/middleware/validate"
// 	"github.com/go-kratos/kratos/v2/transport"
// 	"github.com/go-kratos/kratos/v2/transport/http"
// 	"go.opentelemetry.io/otel/propagation"
// 	"golang.org/x/sync/singleflight"
// 	"google.golang.org/grpc"
// 	"tmc-gitlab.trasre.com/im/api-spec/pkg/rpc/auth"
// 	"tmc-gitlab.trasre.com/im/pkg/cloud/config"
// 	"tmc-gitlab.trasre.com/im/pkg/cloud/config/capollo"
// 	metrics2 "tmc-gitlab.trasre.com/im/pkg/cloud/metrics"
// 	"tmc-gitlab.trasre.com/im/pkg/middleware"
// 	"tmc-gitlab.trasre.com/im/pkg/traffic/igrpc"
// 	"tmc-gitlab.trasre.com/im/pkg/utils/common"
// 	"tmc-gitlab.trasre.com/im/pkg/utils/ctxutil"
// 	"tmc-gitlab.trasre.com/im/pkg/utils/env"
// 	"tmc-gitlab.trasre.com/im/pkg/utils/logs"
// )

// type Http struct {
// 	Network string `json:"network,omitempty"`
// 	Addr    string `json:"addr,omitempty"`
// 	Timeout string `json:"timeout,omitempty"`
// }

// type NoAuth struct {
// 	Services []string `json:"service,omitempty"`
// 	UrlPaths []string `json:"api,omitempty"`
// }

// const (
// 	HTTPDefaultAddr    = "0.0.0.0:8080"
// 	HTTPDefaultTimeout = 3000 * time.Millisecond
// )

// var DefaultMiddlewares = []kmid.Middleware{
// 	recovery.Recovery(recovery.WithLogger(logs.GetAccessLog())),
// 	metrics.Server(
// 		metrics.WithSeconds(prom.NewHistogram(metrics2.ServerMetricSeconds)),
// 		metrics.WithRequests(prom.NewCounter(metrics2.ServerMetricRequests)),
// 	),
// 	tracing.Server(
// 		tracing.WithPropagator(propagation.NewCompositeTextMapPropagator(
// 			tracing.Metadata{}, propagation.TraceContext{}, propagation.Baggage{})),
// 	),
// 	middleware.ServerIErrorMiddleware(),
// 	metadata.Server(),
// 	middleware.AccessLogMiddleware(),
// 	validate.Validator(),
// 	ratelimit.Server(),
// 	middleware.CommonHeaderMiddleware(),
// 	AuthenticationMiddleware(),
// }

// var (
// 	sg          singleflight.Group
// 	authRpcConn *grpc.ClientConn
// )

// func AddMiddleware(middlewares ...kmid.Middleware) {
// 	DefaultMiddlewares = append(DefaultMiddlewares, middlewares...)
// }

// func HttpServer(opts ...http.ServerOption) *http.Server {
// 	opts = append([]http.ServerOption{
// 		http.Middleware(DefaultMiddlewares...),
// 		http.ResponseEncoder(middleware.CommonResponseFunc),
// 		http.ErrorEncoder(middleware.CommonErrorEncoder),
// 		http.Logger(log.DefaultLogger),
// 		http.Address(HTTPDefaultAddr),
// 		http.Timeout(HTTPDefaultTimeout),
// 	}, opts...)
// 	var c Http
// 	if err := config.Scan("", "server.http", &c); err != nil {
// 		panic(err)
// 	}
// 	if c.Network != "" {
// 		opts = append(opts, http.Network(c.Network))
// 	}
// 	if c.Addr != "" {
// 		opts = append(opts, http.Address(c.Addr))
// 	}
// 	if c.Timeout != "" {
// 		duration, err := time.ParseDuration(c.Timeout)
// 		if err != nil {
// 			panic(err.Error())
// 		}
// 		opts = append(opts, http.Timeout(duration))
// 	}
// 	srv := http.NewServer(opts...)

// 	srv.HandleFunc("/ping", func(writer netHttp.ResponseWriter, request *netHttp.Request) {
// 		writer.WriteHeader(netHttp.StatusOK)
// 		writer.Write([]byte("pong"))
// 	})

// 	return srv
// }

// func AuthenticationMiddleware() kratosMiddleware.Middleware {
// 	return func(handler kratosMiddleware.Handler) kratosMiddleware.Handler {
// 		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
// 			if env.GetEnv() == env.Local { // 本地环境开发环境不需要认证
// 				reply, err = handler(ctx, req)
// 				return
// 			}
// 			if tr, ok := transport.FromServerContext(ctx); ok {
// 				if tr.Kind() == transport.KindHTTP {
// 					ht, ok := tr.(*http.Transport)
// 					if !ok {
// 						return nil, errors.New("transport is not http")
// 					}
// 					path := ht.Request().URL.Path
// 					srvName := env.ServiceName()
// 					noauth := getNoAuthConfig()
// 					if !common.HasInArray(srvName, noauth.Services) && !common.HasInArray(path, noauth.UrlPaths) { // 需要进行auth认证
// 						uid, err := authentication(ctx, tr.RequestHeader().Get("Token"))
// 						if err != nil {
// 							return nil, err
// 						}
// 						ctx = context.WithValue(ctx, ctxutil.CtxUserIdKey, uid)
// 					}
// 				}
// 			}
// 			reply, err = handler(ctx, req)
// 			return
// 		}
// 	}
// }

// func authentication(ctx context.Context, token string) (int64, error) {
// 	authClient, er, _ := sg.Do("authConn", func() (any, error) {
// 		if authRpcConn == nil {
// 			conn, err := igrpc.NewConn("nxcore-auth")
// 			if err != nil {
// 				panic(err)
// 			}
// 			authRpcConn = conn
// 		}
// 		client := auth.NewAuthRpcV1Client(authRpcConn)
// 		return client, nil
// 	})
// 	if er != nil {
// 		panic(er)
// 	}

// 	resp, err := authClient.(auth.AuthRpcV1Client).Authentication(ctx, &auth.AuthenticationRequest{Token: token})
// 	if err != nil {
// 		return 0, err
// 	}
// 	return resp.AuthUser.Uid, nil
// }

// func getNoAuthConfig() *NoAuth {
// 	noAuth := &NoAuth{
// 		Services: []string{},
// 		UrlPaths: []string{},
// 	}
// 	noAuthService := config.String(capollo.AuthCommonNameSpance, "noAuthService", "")
// 	if noAuthService != "" {
// 		noAuth.Services = strings.Split(noAuthService, ",")
// 	}
// 	noAuthUrlPaths := config.String(capollo.AuthCommonNameSpance, "noAuthUrlPath", "")
// 	if noAuthUrlPaths != "" {
// 		noAuth.UrlPaths = strings.Split(noAuthUrlPaths, ",")
// 	}
// 	return noAuth
// }
