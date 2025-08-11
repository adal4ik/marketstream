package handlers

import (
	driverinterfaces "marketstream/internal/core/ports/driver"
)

type PriceHandler struct {
	BaseHandler
	service driverinterfaces.PriceDriverInterface
}

func NewPriceHandler(baseHandler *BaseHandler, service driverinterfaces.PriceDriverInterface) *PriceHandler {
	return &PriceHandler{
		BaseHandler: *baseHandler,
		service:     service,
	}
}
