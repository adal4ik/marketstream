package exchange

import (
	"context"
	"encoding/json"
	"math/rand"
	"time"
)

type GenTick struct {
	Symbol    string  `json:"symbol"`
	Price     float64 `json:"price"`
	Timestamp int64   `json:"timestamp"`
}

// Generator пишет в out такой же JSON, как настоящие источники.
type Generator struct {
	Name  string
	Pairs []string
	Hz    int // тиков в секунду на пару (минимум 1)
}

func (g *Generator) Run(ctx context.Context, out chan<- []byte) {
	if g.Hz <= 0 {
		g.Hz = 5
	}
	// сглаженный рандом-бродячий процесс
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	base := map[string]float64{
		"BTCUSDT":  100000,
		"ETHUSDT":  3000,
		"SOLUSDT":  200,
		"DOGEUSDT": 0.3,
		"TONUSDT":  3.5,
	}
	step := map[string]float64{
		"BTCUSDT":  200,
		"ETHUSDT":  30,
		"SOLUSDT":  4,
		"DOGEUSDT": 0.01,
		"TONUSDT":  0.05,
	}
	// инициализация
	state := make(map[string]float64, len(g.Pairs))
	for _, p := range g.Pairs {
		if v, ok := base[p]; ok {
			state[p] = v
		} else {
			state[p] = 100 // дефолт
		}
	}

	interval := time.Second / time.Duration(g.Hz)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case t := <-ticker.C:
			ts := t.UTC().UnixMilli()
			for _, p := range g.Pairs {
				d := step[p]
				if d == 0 {
					d = 1
				}
				// случайный дрейф
				state[p] += (r.Float64()*2 - 1) * d
				if state[p] < 0 {
					state[p] = d
				}
				msg := GenTick{Symbol: p, Price: state[p], Timestamp: ts}
				b, _ := json.Marshal(msg)
				out <- b
			}
		}
	}
}
