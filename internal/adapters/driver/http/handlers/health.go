package handlers

import (
	"net/http"

	"marketstream/internal/core/service"
)

type HealthCheckHandler struct {
	BaseHandler
	Service *service.HealthCheckService
}

func NewHealthCheckHandler(baseHandler *BaseHandler, service *service.HealthCheckService) *HealthCheckHandler {
	return &HealthCheckHandler{BaseHandler: *baseHandler, Service: service}
}

func (h *HealthCheckHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	if err := h.Service.Check(); err != nil {
		h.handleError(w, r, http.StatusServiceUnavailable, "Service Unavailable", err)
	}
	w.Write([]byte("OK"))
}
