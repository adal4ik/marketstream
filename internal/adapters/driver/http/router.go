package http

import (
	"marketstream/internal/adapters/driver/http/handlers"
	"net/http"
)

func NewRouter(handlers *handlers.Handlers) *http.ServeMux {
	mux := http.NewServeMux()

	// // Market Data API
	// mux.HandleFunc("GET /prices/latest/{symbol}", handlers.GetLatestPrice)                      // GET /prices/latest/{symbol}
	// mux.HandleFunc("GET /prices/latest/{exchange}/{symbol}", handlers.GetLatestPriceByExchange) // GET /prices/latest/{exchange}/{symbol}

	// mux.HandleFunc("GET /prices/highest/{symbol}", handlers.GetHighestPrice)                                              // GET /prices/highest/{symbol}
	// mux.HandleFunc("GET /prices/highest/{exchange}/{symbol}", handlers.GetHighestPriceByExchange)                         // GET /prices/highest/{exchange}/{symbol}
	// mux.HandleFunc("GET /prices/highest/{symbol}?period={duration}", handlers.GetHighestPricePeriod)                      // GET /prices/highest/{symbol}?period={duration}
	// mux.HandleFunc("GET /prices/highest/{exchange}/{symbol}?period={duration}", handlers.GetHighestPriceByExchangePeriod) // GET /prices/highest/{exchange}/{symbol}?period={duration}

	// mux.HandleFunc("GET /prices/lowest/{symbol}", handlers.GetLowestPrice)                                              // GET /prices/lowest/{symbol}
	// mux.HandleFunc("GET /prices/lowest/{exchange}/{symbol}", handlers.GetLowestPriceByExchange)                         // GET /prices/lowest/{exchange}/{symbol}
	// mux.HandleFunc("GET /prices/lowest/{symbol}?period={duration}", handlers.GetLowestPricePeriod)                      // GET /prices/lowest/{symbol}?period={duration}
	// mux.HandleFunc("GET /prices/lowest/{exchange}/{symbol}?period={duration}", handlers.GetLowestPriceByExchangePeriod) // GET /prices/lowest/{exchange}/{symbol}?period={duration}

	// mux.HandleFunc("GET /prices/average/{symbol}", handlers.GetAveragePrice)                                              // GET /prices/average/{symbol}
	// mux.HandleFunc("GET /prices/average/{exchange}/{symbol}", handlers.GetAveragePriceByExchange)                         // GET /prices/average/{exchange}/{symbol}
	// mux.HandleFunc("GET /prices/average/{exchange}/{symbol}?period={duration}", handlers.GetAveragePriceByExchangePeriod) // GET /prices/average/{exchange}/{symbol}?period={duration}

	// // Data Mode API
	// mux.HandleFunc("POST /mode/test", handlers.SwitchToTestMode) // POST /mode/test
	// mux.HandleFunc("POST /mode/live", handlers.SwitchToLiveMode) // POST /mode/live

	// // System Health
	// mux.HandleFunc("GET /health", handlers.HealthCheck) // GET /health

	return mux
}
