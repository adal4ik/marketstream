package handlers

type ModeHandler struct {
	BaseHandler
}

func NewModeHandler(baseHandler BaseHandler) *ModeHandler {
	return &ModeHandler{BaseHandler: baseHandler}
}
