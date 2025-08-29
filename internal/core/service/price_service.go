package service

import (
	"context"
	"strconv"
	"strings"
	"time"

	"marketstream/internal/utils"

	"github.com/redis/go-redis/v9"
)

type PriceService struct {
	rdb       *redis.Client
	exchanges []string
}

func NewPriceService(rdb *redis.Client, exchanges []string) *PriceService {
	norm := make([]string, 0, len(exchanges))
	for _, e := range exchanges {
		norm = append(norm, strings.ToLower(e))
	}
	return &PriceService{rdb: rdb, exchanges: norm}
}

func (s *PriceService) Latest(ctx context.Context, symbol string) (price float64, ok bool, err error) {
	key := "latest:" + utils.UpperASCII(symbol)
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

func (s *PriceService) LatestByExchange(ctx context.Context, exchange, symbol string) (price float64, ok bool, err error) {
	ex := strings.ToLower(exchange)
	key := "latest:" + ex + ":" + utils.UpperASCII(symbol)
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

func (s *PriceService) Stats(ctx context.Context, exchange, symbol string, period time.Duration) (min, max, avg float64, ok bool, err error) {
	now := time.Now().UTC()
	from := now.Add(-period)
	fromMs := from.UnixMilli()
	toMs := now.UnixMilli()

	if exchange != "" {
		ex := strings.ToLower(exchange)
		mi, ma, sum, cnt, okOne, e := s.statsOne(ctx, ex, symbol, fromMs, toMs)
		if e != nil || !okOne {
			return 0, 0, 0, false, e
		}
		return mi, ma, sum / float64(cnt), true, nil
	}

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

func (s *PriceService) statsOne(ctx context.Context, exchange, symbol string, fromMs, toMs int64) (min, max, sum float64, count int, ok bool, err error) {
	key := "ticks:" + strings.ToLower(exchange) + ":" + utils.UpperASCII(symbol)
	res, e := s.rdb.ZRangeByScoreWithScores(ctx, key, &redis.ZRangeBy{
		Min:    strconv.FormatInt(fromMs, 10),
		Max:    strconv.FormatInt(toMs, 10),
		Offset: 0,
		Count:  0,
	}).Result()
	if e != nil {
		return 0, 0, 0, 0, false, e
	}
	if len(res) == 0 {
		return 0, 0, 0, 0, false, nil
	}

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
