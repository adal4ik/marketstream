package web

import (
	"net/http"

	"marketstream/internal/adapters/driver/web/handlers"
)

func NewRouter(handlers handlers.Handlers) *http.ServeMux {
	mux := http.NewServeMux()

	return mux
}
