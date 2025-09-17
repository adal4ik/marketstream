package service

import (
	"context"
	"log/slog"
	"strings"
	"sync"
)

type Mode int

const (
	ModeLive Mode = iota
	ModeTest
)

func ParseMode(s string) Mode {
	switch s {
	case "test", "TEST", "Test":
		return ModeTest
	default:
		return ModeLive
	}
}

type ModeService struct {
	log     *slog.Logger
	sources *SourceManager
	prices  *PriceService

	modeMu sync.RWMutex
	mode   Mode

	// текущее состояние конфигурации режимов
	liveAddrs []string
	testNum   int
	testHz    int
	// testPairs можно хранить при желании; для идемпотентности обычно num/hz хватает

	switchMu  sync.Mutex
	srcCtx    context.Context
	srcCancel context.CancelFunc
}

func NewModeService(log *slog.Logger, sm *SourceManager, prices *PriceService, initial Mode) *ModeService {
	return &ModeService{
		log:     log,
		sources: sm,
		prices:  prices,
		mode:    initial,
	}
}

func (m *ModeService) currentMode() Mode {
	m.modeMu.RLock()
	defer m.modeMu.RUnlock()
	return m.mode
}

func (m *ModeService) setMode(md Mode) {
	m.modeMu.Lock()
	m.mode = md
	m.modeMu.Unlock()
}

// --- публичные проверки для хендлеров (чтобы «запретить» лишнюю операцию) ---

func (m *ModeService) IsLiveWith(addrs []string) bool {
	m.modeMu.RLock()
	defer m.modeMu.RUnlock()
	return m.mode == ModeLive && sameStrings(m.liveAddrs, addrs)
}

func (m *ModeService) IsTestWith(num, hz int) bool {
	m.modeMu.RLock()
	defer m.modeMu.RUnlock()
	// если не задано — применяем дефолты, чтобы сравнение было честным
	if num <= 0 {
		num = 3
	}
	if hz <= 0 {
		hz = 5
	}
	return m.mode == ModeTest && m.testNum == num && m.testHz == hz
}

// --- внутреннее ---

func (m *ModeService) shutdownSources() {
	if m.srcCancel != nil {
		m.srcCancel()
		m.srcCancel = nil
		m.srcCtx = nil
	}
	if m.sources != nil {
		m.sources.StopAll()
	}
}

func sameStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	cnt := make(map[string]int, len(a))
	for _, x := range a {
		cnt[strings.ToLower(x)]++
	}
	for _, y := range b {
		k := strings.ToLower(y)
		if cnt[k] == 0 {
			return false
		}
		cnt[k]--
	}
	for _, v := range cnt {
		if v != 0 {
			return false
		}
	}
	return true
}

// --- переключения ---

func (m *ModeService) SwitchToLive(parent context.Context, addrs []string) {
	m.switchMu.Lock()
	defer m.switchMu.Unlock()

	m.shutdownSources()

	m.srcCtx, m.srcCancel = context.WithCancel(parent)

	// имена бирж для агрегаций «по всем биржам»
	exNames := make([]string, len(addrs))
	for i := range addrs {
		exNames[i] = "exchange" + itoa(i+1)
	}

	m.prices.SetExchanges([]string{"exchange1", "exchange2", "exchange3"})

	// запоминаем текущую live-конфигурацию
	m.liveAddrs = append([]string(nil), addrs...)
	m.testNum, m.testHz = 0, 0 // сброс test-конфига

	m.sources.StartLive(m.srcCtx, addrs)

	m.setMode(ModeLive)
	m.log.Info("mode switched to live", "exchanges", exNames)
}

func (m *ModeService) SwitchToTest(parent context.Context, num int, hz int, pairs []string) {
	m.switchMu.Lock()
	defer m.switchMu.Unlock()

	if num <= 0 {
		num = 3
	}
	if hz <= 0 {
		hz = 5
	}

	m.shutdownSources()

	m.srcCtx, m.srcCancel = context.WithCancel(parent)

	exNames := make([]string, 0, num)
	for i := 1; i <= num; i++ {
		exNames = append(exNames, "testex"+itoa(i))
	}

	m.prices.SetExchanges(exNames)

	// запоминаем текущую test-конфигурацию
	m.testNum, m.testHz = num, hz
	m.liveAddrs = nil // сброс live-конфига

	m.sources.StartTest(m.srcCtx, num, hz)

	m.setMode(ModeTest)
	m.log.Info("mode switched to test", "exchanges", exNames, "hz", hz)
}

// локальная itoa без strconv
func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var buf [20]byte
	pos := len(buf)
	n := i
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		pos--
		buf[pos] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}
