package handlers

import (
	"context"
	"encoding/json"
	"marketstream/internal/core/service"
	"net/http"
	"time"
)

type HealthCheckHandler struct {
	BaseHandler
	Service *service.HealthCheckService
}

func NewHealthCheckHandler(baseHandler *BaseHandler, svc *service.HealthCheckService) *HealthCheckHandler {
	return &HealthCheckHandler{BaseHandler: *baseHandler, Service: svc}
}

func (h *HealthCheckHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	// общий таймаут на health (чтобы не висеть дольше секунды)
	ctx, cancel := context.WithTimeout(r.Context(), time.Second)
	defer cancel()

	status := h.Service.Check(ctx)

	code := http.StatusOK
	if status.Status != "ok" {
		code = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(status)
}
