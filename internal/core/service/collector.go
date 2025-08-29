package service

import (
	"context"
	"log/slog"
	"sync/atomic"
	"time"
)

type Collector struct {
	log      *slog.Logger
	received atomic.Int64
}

func NewCollector(log *slog.Logger) *Collector {
	return &Collector{log: log}
}

// Run дренит fan-in канал и безопасно считает тики.
// Это предотвращает блокировки при переполнении буфера.
func (c *Collector) Run(ctx context.Context, in <-chan Tick) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case t := <-in:
			_ = t // тут можно делать что-то полезное (метрики/трейс/батч и т.п.)
			c.received.Add(1)

		case <-ticker.C:
			n := c.received.Swap(0)
			if n > 0 {
				c.log.Info("ticks collected", "count", n)
			}
		}
	}
}
