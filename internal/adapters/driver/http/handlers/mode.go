package handlers

import (
	"context"
	"net/http"
	"strconv"

	"marketstream/internal/core/service"
)

type ModeHandler struct {
	appctx context.Context
	BaseHandler
	Svc       *service.ModeService
	Pairs     []string
	LiveAddrs []string
}

func NewModeHandler(base *BaseHandler, svc *service.ModeService, pairs []string, liveAddrs []string, appctx context.Context) *ModeHandler {
	return &ModeHandler{BaseHandler: *base, Svc: svc, Pairs: pairs, LiveAddrs: liveAddrs, appctx: appctx}
}

func (h *ModeHandler) Live(w http.ResponseWriter, r *http.Request) {
	// запрет: если уже live с теми же адресами — вернуть 409
	if h.Svc.IsLiveWith(h.LiveAddrs) {
		h.handleError(w, r, http.StatusConflict, "already in live mode with the same configuration", nil)
		return
	}

	h.Svc.SwitchToLive(h.appctx, h.LiveAddrs)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok","mode":"live"}`))
}

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

	// запрет: если уже test с теми же параметрами — вернуть 409
	if h.Svc.IsTestWith(num, hz) {
		h.handleError(w, r, http.StatusConflict, "already in test mode with the same configuration", nil)
		return
	}

	h.Svc.SwitchToTest(h.appctx, num, hz, h.Pairs)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok","mode":"test"}`))
}
