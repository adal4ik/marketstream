package handlers

import (
	"log/slog"
	"net/http"

	"marketstream/internal/utils"
)

type BaseHandler struct {
	logger slog.Logger
}

func NewBaseHandler(logger slog.Logger) *BaseHandler {
	return &BaseHandler{
		logger: logger,
	}
}

func (b *BaseHandler) handleError(w http.ResponseWriter, r *http.Request, code int, message string, err error) {
	if err != nil {
		b.logger.Error(message, "error", err, "code", code, "url", r.URL.Path)
	} else {
		b.logger.Error(message, "code", code, "url", r.URL.Path)
	}

	jsonErr := utils.APIError{
		Code:     code,
		Message:  message,
		Resource: r.URL.Path,
	}
	jsonErr.Send(w)
}

type Handlers struct {
	HealthCheck  *HealthCheckHandler
	ModeHandler  *ModeHandler
	PriceHandler *PriceHandler
}

func New(baseHandler BaseHandler) *Handlers {
	return &Handlers{
		HealthCheck:  NewHealthCheckHandler(baseHandler),
		ModeHandler:  NewModeHandler(baseHandler),
		PriceHandler: NewPriceHandler(baseHandler),
	}
}
