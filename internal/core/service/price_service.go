package service

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type PriceService struct {
	rdb       *redis.Client
	exchanges []string
}

func NewPriceService(rdb *redis.Client, exchanges []string) *PriceService {
	// ожидаем имена бирж в lowercase: ["exchange1","exchange2","exchange3"]
	// на всякий случай нормализуем
	norm := make([]string, 0, len(exchanges))
	for _, e := range exchanges {
		norm = append(norm, strings.ToLower(e))
	}
	return &PriceService{rdb: rdb, exchanges: norm}
}

// latest:* (без биржи)
func (s *PriceService) Latest(ctx context.Context, symbol string) (price float64, ok bool, err error) {
	key := "latest:" + upper(symbol)
	val, e := s.rdb.Get(ctx, key).Result()
	if e == redis.Nil {
		return 0, false, nil
	}
	if e != nil {
		return 0, false, e
	}
	p, perr := strconv.ParseFloat(val, 64)
	return p, perr == nil, perr
}

// latest:{exchange}:{symbol}
func (s *PriceService) LatestByExchange(ctx context.Context, exchange, symbol string) (price float64, ok bool, err error) {
	ex := strings.ToLower(exchange)
	key := "latest:" + ex + ":" + upper(symbol)
	val, e := s.rdb.Get(ctx, key).Result()
	if e == redis.Nil {
		return 0, false, nil
	}
	if e != nil {
		return 0, false, e
	}
	p, perr := strconv.ParseFloat(val, 64)
	return p, perr == nil, perr
}

// Stats за период [now-period, now].
// Если exchange == "" — объединяем по всем биржам (из конфигурации) с корректным avg.
func (s *PriceService) Stats(ctx context.Context, exchange, symbol string, period time.Duration) (min, max, avg float64, ok bool, err error) {
	now := time.Now().UTC()
	from := now.Add(-period)
	fromMs := from.UnixMilli()
	toMs := now.UnixMilli()

	// одна биржа
	if exchange != "" {
		ex := strings.ToLower(exchange)
		mi, ma, sum, cnt, okOne, e := s.statsOne(ctx, ex, symbol, fromMs, toMs)
		if e != nil || !okOne {
			return 0, 0, 0, false, e
		}
		return mi, ma, sum / float64(cnt), true, nil
	}

	// все биржи
	totalSum := 0.0
	totalCnt := 0
	first := true

	for _, ex := range s.exchanges {
		mi, ma, sum, cnt, okOne, e := s.statsOne(ctx, ex, symbol, fromMs, toMs)
		if e != nil || !okOne {
			continue
		}
		if first {
			min, max = mi, ma
			first = false
		} else {
			if mi < min {
				min = mi
			}
			if ma > max {
				max = ma
			}
		}
		totalSum += sum
		totalCnt += cnt
		ok = true
	}
	if !ok || totalCnt == 0 {
		return 0, 0, 0, false, nil
	}
	avg = totalSum / float64(totalCnt)
	return min, max, avg, true, nil
}

// одна биржа: возвращаем min, max, sum и count (для корректного объединения)
func (s *PriceService) statsOne(ctx context.Context, exchange, symbol string, fromMs, toMs int64) (min, max, sum float64, count int, ok bool, err error) {
	key := "ticks:" + strings.ToLower(exchange) + ":" + upper(symbol)
	res, e := s.rdb.ZRangeByScoreWithScores(ctx, key, &redis.ZRangeBy{
		Min:    strconv.FormatInt(fromMs, 10),
		Max:    strconv.FormatInt(toMs, 10),
		Offset: 0,
		Count:  0, // всё окно
	}).Result()
	if e != nil {
		return 0, 0, 0, 0, false, e
	}
	if len(res) == 0 {
		return 0, 0, 0, 0, false, nil
	}

	// converter
	toF := func(v any) (float64, bool) {
		switch t := v.(type) {
		case float64:
			return t, true
		case string:
			f, e := strconv.ParseFloat(t, 64)
			return f, e == nil
		case []byte:
			f, e := strconv.ParseFloat(string(t), 64)
			return f, e == nil
		default:
			return 0, false
		}
	}

	// инициализация
	firstVal, okv := toF(res[0].Member)
	if !okv {
		return 0, 0, 0, 0, false, nil
	}
	min, max, sum = firstVal, firstVal, 0.0
	count = 0

	for _, z := range res {
		val, okV := toF(z.Member)
		if !okV {
			continue
		}
		if val < min {
			min = val
		}
		if val > max {
			max = val
		}
		sum += val
		count++
	}
	return min, max, sum, count, true, nil
}

func upper(s string) string {
	// ASCII upper
	b := []byte(s)
	for i := range b {
		if b[i] >= 'a' && b[i] <= 'z' {
			b[i] -= 'a' - 'A'
		}
	}
	return string(b)
}
