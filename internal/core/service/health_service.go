package service

import (
	"context"
	"database/sql"
	"time"

	"github.com/redis/go-redis/v9"
)

type HealthComponent struct {
	OK        bool    `json:"ok"`
	LatencyMs float64 `json:"latency_ms"`
	Error     string  `json:"error,omitempty"`
}

type HealthStatus struct {
	Status   string          `json:"status"` // "ok" | "degraded" | "down"
	Postgres HealthComponent `json:"postgres"`
	Redis    HealthComponent `json:"redis"`
}

type HealthCheckService struct {
	db  *sql.DB
	rdb *redis.Client
}

func NewHealthCheckService(db *sql.DB, rdb *redis.Client) *HealthCheckService {
	return &HealthCheckService{db: db, rdb: rdb}
}

// Check проверяет PG и Redis с таймаутом на каждый компонент.
// Возвращает подробный статус и итоговый агрегированный статус.
func (h *HealthCheckService) Check(ctx context.Context) HealthStatus {
	const perCheckTimeout = 300 * time.Millisecond

	res := HealthStatus{
		Status:   "ok",
		Postgres: HealthComponent{},
		Redis:    HealthComponent{},
	}

	// --- Postgres ---
	func() {
		cctx, cancel := context.WithTimeout(ctx, perCheckTimeout)
		defer cancel()

		start := time.Now()
		if h.db == nil {
			res.Postgres.OK = false
			res.Postgres.Error = "db is nil"
			return
		}
		if err := h.db.PingContext(cctx); err != nil {
			res.Postgres.OK = false
			res.Postgres.Error = err.Error()
			return
		}
		res.Postgres.OK = true
		res.Postgres.LatencyMs = float64(time.Since(start).Milliseconds())
	}()

	// --- Redis ---
	func() {
		cctx, cancel := context.WithTimeout(ctx, perCheckTimeout)
		defer cancel()

		start := time.Now()
		if h.rdb == nil {
			// допускаем режим без Redis: считаем degraded, а не down
			res.Redis.OK = false
			res.Redis.Error = "redis is nil"
			return
		}
		if err := h.rdb.Ping(cctx).Err(); err != nil {
			res.Redis.OK = false
			res.Redis.Error = err.Error()
			return
		}
		res.Redis.OK = true
		res.Redis.LatencyMs = float64(time.Since(start).Milliseconds())
	}()

	// агрегированный статус
	switch {
	case res.Postgres.OK && res.Redis.OK:
		res.Status = "ok"
	case res.Postgres.OK || res.Redis.OK:
		res.Status = "degraded"
	default:
		res.Status = "down"
	}

	return res
}
