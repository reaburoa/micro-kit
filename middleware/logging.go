package middleware

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/welltop-cn/common/utils/log"
)

// 网关传递请求header头：
// X-Forwarded-Host Remoteip Authuser Trace_flag X-Real-Ip X-Forwarded-For Accept Grpc-Timeout Traceparent X-Scheme X-Original-Forwarded-For User-Agent Cache-Control Postman-Token Accept-Encoding Span_id X-Forwarded-Scheme Token X-Forwarded-Proto Trace_id X-Request-Id X-Forwarded-Port
func AccessLogMiddleware() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			start := time.Now()
			var (
				code      int32
				message   string
				accessLog = make(map[string]interface{}, 10)
			)
			if tr, ok := transport.FromServerContext(ctx); ok {
				userIp := getIp(tr.RequestHeader())
				accessLog = map[string]interface{}{
					"proto":      tr.Operation(),
					"user-agent": getUa(tr.RequestHeader()),
					"remote":     userIp,
					//"app-version": ctxutil.GetCommonHeader(ctx, ctxutil.CommonHeaderAppVersion),
					"trace_id":   tr.RequestHeader().Get("Trace_id"),
					"user-token": tr.RequestHeader().Get("Token"),
					"X-Scheme":   tr.RequestHeader().Get("X-Scheme"),
					"kind":       tr.Kind().String(),
					"endpoint":   tr.Endpoint(),
					"kind_type":  "server",
				}
				//ctx = context.WithValue(ctx, ctxutil.CtxUserIpKey, userIp)
			}
			reply, err = handler(ctx, req)
			// if err != nil {
			// 	code = ierrors.Code(err)
			// 	message = ierrors.Msg(err)
			// }
			accessLog["args"] = extractArgs(req)
			accessLog["respCode"] = code
			accessLog["respMsg"] = message
			// if accessLog["trace_id"] == "" {
			// 	accessLog["trace_id"] = ctxutil.GetTraceID(ctx)
			// }
			defer func(begin time.Time) {
				accessLog["request_time"] = time.Since(begin).Seconds()
				log.CtxInfof(ctx, "accessLog %#v", accessLog)
			}(start)
			return
		}
	}
}

func ClientCallLogMiddleware() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			start := time.Now()
			var (
				code    int32
				message string
				callLog = make(map[string]interface{}, 10)
			)
			if tr, ok := transport.FromClientContext(ctx); ok {
				callLog = map[string]interface{}{
					"proto":     tr.Operation(),
					"kind":      tr.Kind().String(),
					"endpoint":  tr.Endpoint(),
					"kind_type": "client",
					//"trace_id":  ctxutil.GetTraceID(ctx),
				}
			}
			reply, err = handler(ctx, req)
			// if err != nil {
			// 	code = ierrors.Code(err)
			// 	message = ierrors.Msg(err)
			// }
			callLog["args"] = extractArgs(req)
			callLog["respCode"] = code
			callLog["respMsg"] = message
			defer func(begin time.Time) {
				callLog["request_time"] = time.Since(begin).Seconds()
				log.CtxInfof(ctx, "callLog %#v", callLog)
			}(start)

			return
		}
	}
}

// extractArgs returns the string of the req
func extractArgs(req interface{}) string {
	if stringer, ok := req.(fmt.Stringer); ok {
		return stringer.String()
	}
	return fmt.Sprintf("%+v", req)
}

func getUa(trHeader transport.Header) string {
	ua := trHeader.Get("User-Agent")
	if ua == "" {
		ua = trHeader.Get("user-agent")
	}

	return ua
}

// 通过网关Header获取IP地址
func getIp(trHeader transport.Header) string {
	remoteIp := trHeader.Get("X-Forwarded-For")
	if remoteIp == "" {
		remoteIp = trHeader.Get("X-Real-Ip")
		if remoteIp != "" {
			remoteIp = trHeader.Get("Remoteip")
		}
	}

	return strings.Split(remoteIp, ",")[0]
}
