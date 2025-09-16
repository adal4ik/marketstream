package service

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type Tick struct {
	Exchange  string  `json:"-"`
	Symbol    string  `json:"symbol"`
	Price     float64 `json:"price"`
	Timestamp int64   `json:"timestamp"` // ms
}

type ExchangeService struct {
	rdb     *redis.Client
	allowed map[string]struct{}
}

func NewExchangeService(rdb *redis.Client, pairs []string) *ExchangeService {
	m := make(map[string]struct{}, len(pairs))
	for _, p := range pairs {
		p = strings.ToUpper(strings.TrimSpace(p))
		if p != "" {
			m[p] = struct{}{}
		}
	}
	return &ExchangeService{rdb: rdb, allowed: m}
}

func (s *ExchangeService) isAllowed(sym string) bool {
	_, ok := s.allowed[strings.ToUpper(sym)]
	return ok
}

// Worker: читает сырые JSON-байты, валидирует, пишет в Redis (latest + окно), пробрасывает Tick в fan-in.
func (s *ExchangeService) Worker(ctx context.Context, exName string, in <-chan []byte, out chan<- Tick) {
	for {
		select {
		case <-ctx.Done():
			return
		case raw := <-in:
			var t Tick
			if err := json.Unmarshal(raw, &t); err != nil {
				continue
			}
			if !s.isAllowed(t.Symbol) {
				continue
			}
			if t.Timestamp < 1_000_000_000_000 { // если пришло в секундах
				t.Timestamp *= 1000
			}

			// --- Redis: latest 60s ---
			_ = s.rdb.Set(ctx, "latest:"+strings.ToUpper(t.Symbol), t.Price, 60*time.Second).Err()
			_ = s.rdb.Set(ctx, "latest:"+exName+":"+strings.ToUpper(t.Symbol), t.Price, 60*time.Second).Err()

			// --- Redis: окно 60s с ZSET ---
			key := "ticks:" + exName + ":" + strings.ToUpper(t.Symbol)
			_ = s.rdb.ZAdd(ctx, key, redis.Z{Score: float64(t.Timestamp), Member: t.Price}).Err()
			_ = s.rdb.ZRemRangeByScore(ctx, key, "-inf", fmtFloat(t.Timestamp-60_000)).Err()

			t.Exchange = exName
			select {
			case <-ctx.Done():
				return
			case out <- t:
			}
		}
	}
}

func fmtFloat(v int64) string { return strconv.FormatFloat(float64(v), 'f', -1, 64) }
