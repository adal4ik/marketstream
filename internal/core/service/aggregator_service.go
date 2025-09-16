package service

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"time"

	"marketstream/internal/adapters/driven/database/repository"

	"github.com/redis/go-redis/v9"
)

// динамический провайдер списка бирж
type exchangeProvider interface {
	Exchanges() []string
}

// Простой минутный агрегатор: раз в минуту берёт 60с из Redis по всем актуальным биржам и пишет в PG.
type Aggregator struct {
	rdb    *redis.Client
	repo   *repository.AggregateRepository
	pairs  []string
	exprov exchangeProvider // ← берём биржи не статически, а каждый раз
}

func NewAggregator(rdb *redis.Client, repo *repository.AggregateRepository, pairs []string, exprov exchangeProvider) *Aggregator {
	return &Aggregator{
		rdb:    rdb,
		repo:   repo,
		pairs:  pairs,
		exprov: exprov,
	}
}

func (a *Aggregator) Run(ctx context.Context) {
	if a.rdb == nil || a.repo == nil {
		return
	}

	// тикать будем раз в секунду, но агрегировать один раз в новую минуту (на границе)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	var lastMinute int64 = -1

	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			minKey := now.UTC().Unix() / 60
			if minKey == lastMinute {
				continue
			}
			lastMinute = minKey

			fromMs := now.UTC().Add(-time.Minute).UnixMilli()
			toMs := now.UTC().UnixMilli()

			exchanges := a.exprov.Exchanges() // ← актуальный список (live или test)
			if len(exchanges) == 0 {
				continue
			}

			var recs []repository.MinuteAggregateRecord

			for _, ex := range exchanges {
				for _, pair := range a.pairs {
					key := "ticks:" + ex + ":" + pair
					members, err := a.rdb.ZRangeByScoreWithScores(ctx, key, &redis.ZRangeBy{
						Min:    fmt.Sprintf("%d", fromMs),
						Max:    fmt.Sprintf("%d", toMs),
						Offset: 0,
						Count:  0,
					}).Result()
					if err != nil || len(members) == 0 {
						continue
					}

					var sum float64
					minv := math.MaxFloat64
					maxv := -math.MaxFloat64

					for _, m := range members {
						var price float64
						switch v := m.Member.(type) {
						case string:
							price, _ = strconv.ParseFloat(v, 64)
						case []byte:
							price, _ = strconv.ParseFloat(string(v), 64)
						case float64:
							price = v
						default:
							continue
						}
						sum += price
						if price < minv {
							minv = price
						}
						if price > maxv {
							maxv = price
						}
					}

					avg := sum / float64(len(members))
					recs = append(recs, repository.MinuteAggregateRecord{
						PairName:     pair,
						Exchange:     ex,
						Timestamp:    now.UTC().Truncate(time.Minute), // конец минуты
						AveragePrice: avg,
						MinPrice:     minv,
						MaxPrice:     maxv,
					})
				}
			}

			if len(recs) == 0 {
				continue
			}
			_ = a.repo.BatchInsert(ctx, recs)
		}
	}
}
