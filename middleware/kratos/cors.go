package kratos

import (
	"net/http"

	"github.com/gorilla/handlers"
)

func CORSFilter() func(http.Handler) http.Handler {
	return handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"POST", "GET", "PUT", "DELETE", "HEAD", "UPDATE"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
	)
}
