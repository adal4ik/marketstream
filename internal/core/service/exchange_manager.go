// internal/core/service/exchange_manager.go
package service

import (
	"context"
	"sync"

	"marketstream/internal/adapters/driven/exchange"
)

type Manager struct {
	svc      *ExchangeService
	resultCh chan Tick

	rawChs map[string]chan []byte
	wg     sync.WaitGroup
}

func NewManager(svc *ExchangeService, resultBuf int) *Manager {
	return &Manager{
		svc:      svc,
		resultCh: make(chan Tick, resultBuf),
		rawChs:   make(map[string]chan []byte),
	}
}

func (m *Manager) ResultCh() <-chan Tick { return m.resultCh }

func (m *Manager) StartExchange(ctx context.Context, name, addr string) {
	raw := make(chan []byte, 1024)
	m.rawChs[name] = raw

	// listener (функция, а не тип)
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		// блокирующий вызов: читает TCP и пишет строки в raw;
		// внутри Listen должен быть свой реконнект/выход по ctx.Done()
		exchange.Listen(ctx, addr, raw)
	}()

	// 5 workers
	for i := 0; i < 5; i++ {
		m.wg.Add(1)
		go func() {
			defer m.wg.Done()
			m.svc.Worker(ctx, name, raw, m.resultCh)
		}()
	}
}

func (m *Manager) Wait() { m.wg.Wait() }
