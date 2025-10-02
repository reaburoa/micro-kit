package kratos

import (
	"context"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/go-kratos/kratos/v2/transport/http"
)

// CORSMiddleware 完整的 CORS 中间件
func CORSMiddleware() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			// 如果是 HTTP 请求
			if tr, ok := transport.FromServerContext(ctx); ok {
				if ht, ok := tr.(*http.Transport); ok {
					setCORSHeaders(ht)

					// 处理 Preflight 请求
					if ht.Request().Method == "OPTIONS" {
						return nil, nil
					}
				}
			}

			return handler(ctx, req)
		}
	}
}

// setCORSHeaders 设置 CORS 头
func setCORSHeaders(ht *http.Transport) {
	req := ht.Request()
	headers := ht.ReplyHeader()

	// 设置 CORS 头
	origin := req.Header.Get("Origin")
	if origin == "" {
		origin = "*"
	}

	headers.Set("Access-Control-Allow-Origin", origin)
	headers.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
	headers.Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, X-CSRF-Token, Accept, Origin, Cache-Control, X-Requested-With")
	headers.Set("Access-Control-Allow-Credentials", "true")
	headers.Set("Access-Control-Max-Age", "86400") // 24小时

	// 处理预检请求的额外头
	if reqHeaders := req.Header.Get("Access-Control-Request-Headers"); reqHeaders != "" {
		headers.Set("Access-Control-Allow-Headers", reqHeaders)
	}
}

// func CrosFilter() func(http.Handler) http.Handler {
// 	return handlers.CORS(
// 		handlers.AllowedOrigins([]string{"*"}),
// 		handlers.AllowedMethods([]string{"POST", "GET", "OPTIONS", "PUT", "DELETE", "UPDATE"}),
// 		handlers.AllowedHeaders([]string{"*"}),
// 	)
// }
