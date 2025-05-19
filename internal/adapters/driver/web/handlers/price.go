package handlers

type PriceHandler struct {
	BaseHandler
}

func NewPriceHandler(baseHandler BaseHandler) *PriceHandler {
	return &PriceHandler{BaseHandler: baseHandler}
}
