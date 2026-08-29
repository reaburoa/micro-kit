package kratos

import (
	"net/http"

	"github.com/gorilla/handlers"
	appconfig "github.com/reaburoa/micro-kit/cloud/config"
)

func CORSFilter() func(http.Handler) http.Handler {
	allowedOrigins := []string{"*"}
	allowedMethods := []string{"POST", "GET", "PUT", "DELETE", "HEAD", "OPTIONS", "PATCH", "UPDATE"}
	allowedHeaders := []string{"Content-Type", "Authorization", "X-Request-Id", "Token"}

	if v := appconfig.Get("cors.allowed_origins"); v != nil {
		var origins []string
		if err := v.Scan(&origins); err == nil && len(origins) > 0 {
			allowedOrigins = origins
		}
	}
	if v := appconfig.Get("cors.allowed_methods"); v != nil {
		var methods []string
		if err := v.Scan(&methods); err == nil && len(methods) > 0 {
			allowedMethods = methods
		}
	}
	if v := appconfig.Get("cors.allowed_headers"); v != nil {
		var headers []string
		if err := v.Scan(&headers); err == nil && len(headers) > 0 {
			allowedHeaders = headers
		}
	}

	return handlers.CORS(
		handlers.AllowedOrigins(allowedOrigins),
		handlers.AllowedMethods(allowedMethods),
		handlers.AllowedHeaders(allowedHeaders),
		handlers.AllowCredentials(),
	)
}
