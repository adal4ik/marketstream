package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"marketstream/internal/core/service"
	"marketstream/internal/utils"
)

type PriceHandlers struct {
	Base *BaseHandler
	Svc  *service.PriceService
}

func NewPriceHandlers(base *BaseHandler, svc *service.PriceService) *PriceHandlers {
	return &PriceHandlers{Base: base, Svc: svc}
}

// GET /prices/latest/{symbol}
func (h *PriceHandlers) Latest(w http.ResponseWriter, r *http.Request) {
	symbol := r.PathValue("symbol")
	if symbol == "" {
		h.Base.handleError(w, r, http.StatusBadRequest, "symbol is required", nil)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 300*time.Millisecond)
	defer cancel()

	price, ok, err := h.Svc.Latest(ctx, symbol)
	if err != nil {
		h.Base.handleError(w, r, http.StatusInternalServerError, "redis error", err)
		return
	}
	if !ok {
		h.Base.handleError(w, r, http.StatusNotFound, "no data for symbol", nil)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"symbol": utils.UpperASCII(symbol),
		"price":  price,
	})
}

// GET /prices/latest/{exchange}/{symbol}
func (h *PriceHandlers) LatestByExchange(w http.ResponseWriter, r *http.Request) {
	ex := strings.ToLower(r.PathValue("exchange"))
	symbol := r.PathValue("symbol")
	if ex == "" || symbol == "" {
		h.Base.handleError(w, r, http.StatusBadRequest, "exchange and symbol are required", nil)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 300*time.Millisecond)
	defer cancel()

	price, ok, err := h.Svc.LatestByExchange(ctx, ex, symbol)
	if err != nil {
		h.Base.handleError(w, r, http.StatusInternalServerError, "redis error", err)
		return
	}
	if !ok {
		h.Base.handleError(w, r, http.StatusNotFound, "no data for exchange/symbol", nil)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"exchange": utils.UpperASCII(ex),
		"symbol":   utils.UpperASCII(symbol),
		"price":    price,
	})
}

// GET /prices/highest/{symbol}
// GET /prices/highest/{exchange}/{symbol}?period=...
func (h *PriceHandlers) Highest(w http.ResponseWriter, r *http.Request) {
	h.handleStat(w, r, "highest")
}

// GET /prices/lowest/{symbol}
// GET /prices/lowest/{exchange}/{symbol}?period=...
func (h *PriceHandlers) Lowest(w http.ResponseWriter, r *http.Request) {
	h.handleStat(w, r, "lowest")
}

// GET /prices/average/{symbol}
// GET /prices/average/{exchange}/{symbol}?period=...
func (h *PriceHandlers) Average(w http.ResponseWriter, r *http.Request) {
	h.handleStat(w, r, "average")
}

func (h *PriceHandlers) handleStat(w http.ResponseWriter, r *http.Request, kind string) {
	ex := r.PathValue("exchange")
	if ex != "" {
		ex = strings.ToLower(ex)
	}
	symbol := r.PathValue("symbol")
	if symbol == "" {
		h.Base.handleError(w, r, http.StatusBadRequest, "symbol is required", nil)
		return
	}
	periodStr := r.URL.Query().Get("period")
	if periodStr == "" {
		periodStr = "60s"
	}
	period, err := time.ParseDuration(periodStr)
	if err != nil || period <= 0 || period > 10*time.Minute {
		h.Base.handleError(w, r, http.StatusBadRequest, "invalid period (use 1s..10m, e.g. 30s, 1m, 5m)", nil)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 400*time.Millisecond)
	defer cancel()

	min, max, avg, ok, statErr := h.Svc.Stats(ctx, ex, symbol, period)
	if statErr != nil {
		h.Base.handleError(w, r, http.StatusInternalServerError, "redis error", statErr)
		return
	}
	if !ok {
		h.Base.handleError(w, r, http.StatusNotFound, "no data in window", nil)
		return
	}

	resp := map[string]any{
		"symbol": utils.UpperASCII(symbol),
		"period": period.String(),
	}
	if ex != "" {
		resp["exchange"] = utils.UpperASCII(ex)
	}
	switch kind {
	case "highest":
		resp["value"] = max
	case "lowest":
		resp["value"] = min
	case "average":
		resp["value"] = avg
	}
	writeJSON(w, http.StatusOK, resp)
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
