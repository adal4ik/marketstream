package handlers

type HealthCheckHandler struct {
	BaseHandler
}

func NewHealthCheckHandler(baseHandler BaseHandler) *HealthCheckHandler {
	return &HealthCheckHandler{BaseHandler: baseHandler}
}
