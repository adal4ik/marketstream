package service

import (
	"context"
	"database/sql"

	"github.com/redis/go-redis/v9"
)

type HealthCheckService struct {
	db  *sql.DB
	rdb *redis.Client
}

func NewHealthCheckService(db *sql.DB, rdb *redis.Client) *HealthCheckService {
	return &HealthCheckService{
		db:  db,
		rdb: rdb,
	}
}

func (h *HealthCheckService) Check() error {
	ctx := context.Background()
	if err := h.db.PingContext(ctx); err != nil {
		return err
	}
	if err := h.rdb.Ping(ctx).Err(); err != nil {
		return err
	}
	return nil
}
