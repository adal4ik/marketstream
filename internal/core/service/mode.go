package service

import (
	"context"
	"strings"
	"sync"
)

type Mode string

const (
	ModeLive Mode = "live"
	ModeTest Mode = "test"
)

type ModeService struct {
	mu   sync.Mutex
	mode Mode
	mgr  *SourceManager
}

func NewModeService(mgr *SourceManager, initial Mode) *ModeService {
	return &ModeService{mgr: mgr, mode: initial}
}

func (m *ModeService) Current() Mode {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.mode
}

func (m *ModeService) SwitchToLive(ctx context.Context, exchanges []string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.mgr.StopAll()
	m.mgr.StartLive(ctx, exchanges)
	m.mode = ModeLive
}

func (m *ModeService) SwitchToTest(ctx context.Context, numEx, hz int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.mgr.StopAll()
	m.mgr.StartTest(ctx, numEx, hz)
	m.mode = ModeTest
}

// маленький helper, если из env приходит "LIVE"/"Test" и т.п.
func ParseMode(s string) Mode {
	switch strings.ToLower(s) {
	case "test":
		return ModeTest
	default:
		return ModeLive
	}
}
