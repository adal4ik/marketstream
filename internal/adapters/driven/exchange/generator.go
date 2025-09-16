package exchange

import (
	"context"
	"encoding/json"
	"math/rand"
	"time"
)

// фиксированная цена для тестового режима
const generatorPrice = 100.00

type Generator struct {
	Name  string   // например "testex1"
	Pairs []string // список пар
	Hz    int      // тиков/сек
}

type genTick struct {
	Exchange  string  `json:"exchange"`
	Symbol    string  `json:"symbol"`
	Price     float64 `json:"price"`
	Timestamp int64   `json:"timestamp"` // важно: timestamp (мс), а не time
}

// Run генерит одинаковую цену и пишет JSON в out (chan []byte).
func (g *Generator) Run(ctx context.Context, out chan<- []byte) {
	hz := g.Hz
	if hz <= 0 {
		hz = 5
	}
	period := time.Second / time.Duration(hz)
	t := time.NewTicker(period)
	defer t.Stop()

	time.Sleep(time.Duration(rand.Intn(200)) * time.Millisecond) // небольшой джиттер старта

	for {
		select {
		case <-ctx.Done():
			return
		case now := <-t.C:
			ts := now.UTC().UnixMilli()
			for _, p := range g.Pairs {
				msg := genTick{
					Exchange:  g.Name,
					Symbol:    p,
					Price:     generatorPrice,
					Timestamp: ts, // сразу миллисекунды
				}
				b, _ := json.Marshal(msg)
				select {
				case out <- b:
				case <-ctx.Done():
					return
				}
			}
		}
	}
}
