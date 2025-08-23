package handlers

import (
	"net/http"
	"strconv"

	"marketstream/internal/core/service"
)

type ModeHandler struct {
	Base *BaseHandler
	Svc  *service.ModeService
}

func NewModeHandler(base *BaseHandler, svc *service.ModeService) *ModeHandler {
	return &ModeHandler{Base: base, Svc: svc}
}

// GET /mode
func (h *ModeHandler) Get(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"mode": h.Svc.Current(),
	})
}

// POST /mode/live
func (h *ModeHandler) Live(w http.ResponseWriter, r *http.Request) {
	// exchanges берутся из main (пробрасываются в ModeService.SwitchToLive)
	// здесь просто переключаем
	h.Svc.SwitchToLive(r.Context(), nil) // список уйдёт из main через замыкание или setter (см. ниже)
	writeJSON(w, http.StatusOK, map[string]any{"mode": "live"})
}

// POST /mode/test?num=3&hz=5
func (h *ModeHandler) Test(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	num, _ := strconv.Atoi(q.Get("num"))
	hz, _ := strconv.Atoi(q.Get("hz"))
	if num <= 0 {
		num = 3
	}
	if hz <= 0 {
		hz = 5
	}
	h.Svc.SwitchToTest(r.Context(), num, hz)
	writeJSON(w, http.StatusOK, map[string]any{
		"mode": "test", "num": num, "hz": hz,
	})
}
