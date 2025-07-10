package kratos

import (
	"net/http"

	"github.com/gorilla/handlers"
)

func CrosFilter() func(http.Handler) http.Handler {
	return handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"POST", "GET", "OPTIONS", "PUT", "DELETE", "UPDATE"}),
	)
}
