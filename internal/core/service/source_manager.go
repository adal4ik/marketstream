package service

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"marketstream/internal/adapters/driven/exchange"
)

type SourceManager struct {
	exSvc    *ExchangeService
	pairs    []string
	resultCh chan Tick

	mu        sync.Mutex
	liveStops []context.CancelFunc
	testStops []context.CancelFunc
	wg        sync.WaitGroup
}

func NewSourceManager(exSvc *ExchangeService, pairs []string, resultCh chan Tick) *SourceManager {
	return &SourceManager{
		exSvc:    exSvc,
		pairs:    pairs,
		resultCh: resultCh,
	}
}

func (m *SourceManager) StopAll() {
	m.mu.Lock()
	for _, c := range m.liveStops {
		c()
	}
	for _, c := range m.testStops {
		c()
	}
	m.liveStops = nil
	m.testStops = nil
	m.mu.Unlock()
	// дождаться всех горутин источников/воркеров
	m.wg.Wait()
}

// Live: слушаем TCP источники exchange1..3
func (m *SourceManager) StartLive(ctx context.Context, addrs []string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, addr := range addrs {
		exName := fmt.Sprintf("exchange%d", i+1)
		rawCh := make(chan []byte, 1024)

		// контекст отмены для этого источника
		lctx, cancel := context.WithCancel(ctx)
		m.liveStops = append(m.liveStops, cancel)

		// listener с backoff/reconnect
		m.wg.Add(1)
		go func(name, a string, out chan<- []byte, cctx context.Context) {
			defer m.wg.Done()
			backoffs := []time.Duration{time.Second, 5 * time.Second, 15 * time.Second, 30 * time.Second}
			i := 0
			for {
				select {
				case <-cctx.Done():
					return
				default:
				}
				exchange.Listen(cctx, a, out) // блокирует до разрыва

				wait := backoffs[i]
				if i < len(backoffs)-1 {
					i++
				}
				select {
				case <-cctx.Done():
					return
				case <-time.After(wait):
				}
				_ = name
			}
		}(exName, addr, rawCh, lctx)

		// 5 воркеров
		for w := 0; w < 5; w++ {
			m.wg.Add(1)
			go func(name string, in <-chan []byte, cctx context.Context) {
				defer m.wg.Done()
				m.exSvc.Worker(cctx, name, in, m.resultCh)
			}(exName, rawCh, lctx)
		}
	}
}

// Test: поднимаем локальные генераторы testex1..3
func (m *SourceManager) StartTest(ctx context.Context, numEx int, hz int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if numEx <= 0 {
		numEx = 3
	}
	if hz <= 0 {
		hz = 5
	}

	for i := 1; i <= numEx; i++ {
		exName := fmt.Sprintf("testex%d", i)
		rawCh := make(chan []byte, 1024)

		tctx, cancel := context.WithCancel(ctx)
		m.testStops = append(m.testStops, cancel)

		gen := &exchange.Generator{
			Name:  exName,
			Pairs: m.pairs,
			Hz:    hz,
		}

		// generator
		m.wg.Add(1)
		go func(g *exchange.Generator, out chan<- []byte, cctx context.Context) {
			defer m.wg.Done()
			g.Run(cctx, out)
		}(gen, rawCh, tctx)

		// 5 воркеров
		for w := 0; w < 5; w++ {
			m.wg.Add(1)
			go func(name string, in <-chan []byte, cctx context.Context) {
				defer m.wg.Done()
				// имя биржи в Redis будет lowercase
				m.exSvc.Worker(cctx, strings.ToLower(name), in, m.resultCh)
			}(exName, rawCh, tctx)
		}
	}
}
