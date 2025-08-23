package http

import (
	"marketstream/internal/adapters/driver/http/handlers"
	"net/http"
)

func NewRouter(hs *handlers.Handlers) *http.ServeMux {
	mux := http.NewServeMux()

	// Price API over Redis
	mux.HandleFunc("GET /prices/latest/{symbol}", hs.Price.Latest)
	mux.HandleFunc("GET /prices/latest/{exchange}/{symbol}", hs.Price.LatestByExchange)

	mux.HandleFunc("GET /prices/highest/{symbol}", hs.Price.Highest)
	mux.HandleFunc("GET /prices/highest/{exchange}/{symbol}", hs.Price.Highest)

	mux.HandleFunc("GET /prices/lowest/{symbol}", hs.Price.Lowest)
	mux.HandleFunc("GET /prices/lowest/{exchange}/{symbol}", hs.Price.Lowest)

	mux.HandleFunc("GET /prices/average/{symbol}", hs.Price.Average)
	mux.HandleFunc("GET /prices/average/{exchange}/{symbol}", hs.Price.Average)

	mux.HandleFunc("GET /mode", hs.Mode.Get)
	mux.HandleFunc("POST /mode/live", hs.Mode.Live)
	mux.HandleFunc("POST /mode/test", hs.Mode.Test)

	mux.HandleFunc("GET /health", hs.HealthCheck.HealthCheck)

	return mux
}
