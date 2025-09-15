package service

import (
	"context"
	"marketstream/internal/core/domain"
	"strconv"
	"time"

	drivenports "marketstream/internal/core/ports/driven"

	"github.com/redis/go-redis/v9"
)

type Aggregator struct {
	rdb       *redis.Client
	repo      drivenports.AggregateRepo
	pairs     []domain.Symbol
	exchanges []domain.Exchange
}

func NewAggregator(rdb *redis.Client, repo drivenports.AggregateRepo, pairs []string, exchanges []string) *Aggregator {
	ps := make([]domain.Symbol, 0, len(pairs))
	for _, p := range pairs {
		ps = append(ps, domain.Symbol(p))
	}
	exs := make([]domain.Exchange, 0, len(exchanges))
	for _, e := range exchanges {
		exs = append(exs, domain.Exchange(e))
	}
	return &Aggregator{rdb: rdb, repo: repo, pairs: ps, exchanges: exs}
}

func (a *Aggregator) Run(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	var batch []domain.MinuteAggregate
	lastMinute := time.Now().UTC().Minute()

	for {
		select {
		case <-ctx.Done():
			if len(batch) > 0 {
				_ = a.repo.BatchInsert(context.Background(), batch)
			}
			return

		case now := <-ticker.C:
			now = now.UTC()
			from := now.Add(-time.Minute)
			nowMs, fromMs := now.UnixMilli(), from.UnixMilli()

			for _, ex := range a.exchanges {
				for _, p := range a.pairs {
					min, max, avg, ok, err := a.stats(ctx, string(ex), string(p), fromMs, nowMs)
					if err != nil || !ok {
						continue
					}
					batch = append(batch, domain.MinuteAggregate{
						Pair: p, Exchange: ex, Ts: now,
						Avg: avg, Min: min, Max: max,
					})
				}
			}

			if now.Minute() != lastMinute && len(batch) > 0 {
				_ = a.repo.BatchInsert(ctx, batch)
				batch = batch[:0]
				lastMinute = now.Minute()
			}
		}
	}
}

func (a *Aggregator) stats(ctx context.Context, ex, symbol string, fromMs, toMs int64) (min, max, avg float64, ok bool, err error) {
	key := "ticks:" + ex + ":" + symbol
	res, e := a.rdb.ZRangeByScoreWithScores(ctx, key, &redis.ZRangeBy{
		Min:    strconv.FormatInt(fromMs, 10),
		Max:    strconv.FormatInt(toMs, 10),
		Offset: 0, Count: 5000,
	}).Result()
	if e != nil {
		err = e
		return
	}
	if len(res) == 0 {
		return 0, 0, 0, false, nil
	}

	first, okNum := toFloat(res[0].Member)
	if !okNum {
		return 0, 0, 0, false, nil
	}
	min, max, sum := first, first, 0.0
	for _, z := range res {
		price, okNum := toFloat(z.Member)
		if !okNum {
			continue
		}
		if price < min {
			min = price
		}
		if price > max {
			max = price
		}
		sum += price
	}
	avg = sum / float64(len(res))
	return min, max, avg, true, nil
}

func toFloat(m any) (float64, bool) {
	switch v := m.(type) {
	case float64:
		return v, true
	case string:
		f, err := strconv.ParseFloat(v, 64)
		return f, err == nil
	case []byte:
		f, err := strconv.ParseFloat(string(v), 64)
		return f, err == nil
	default:
		return 0, false
	}
}
