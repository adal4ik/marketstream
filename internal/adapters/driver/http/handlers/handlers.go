package handlers

import (
	"database/sql"
	"log/slog"
	"marketstream/internal/core/service"
	"marketstream/internal/utils"
	"net/http"

	"github.com/redis/go-redis/v9"
)

type BaseHandler struct {
	logger *slog.Logger
}

func NewBaseHandler(logger *slog.Logger) *BaseHandler {
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
	HealthCheck *HealthCheckHandler
	Mode        *ModeHandler
	Price       *PriceHandlers
}

func New(base *BaseHandler, svc *service.Service, db *sql.DB, rdb *redis.Client, pairs []string, liveAddrs []string) *Handlers {
	return &Handlers{
		HealthCheck: NewHealthCheckHandler(base, svc.HelthCheck),
		Mode:        NewModeHandler(base, svc.ModeService, pairs, liveAddrs),
		Price:       NewPriceHandlers(base, svc.PriceService),
	}
}
