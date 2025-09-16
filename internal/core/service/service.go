package service

import (
	"database/sql"
	"log/slog"

	"marketstream/internal/adapters/driven/database/repository"

	"github.com/redis/go-redis/v9"
)

type Service struct {
	HelthCheck   *HealthCheckService
	ModeService  *ModeService
	PriceService *PriceService
	Aggregator   *Aggregator
}

// exchanges: список live-бирж по именам (например: []string{"exchange1","exchange2","exchange3"})
// liveAddrs: адреса tcp для live (например: []string{"exchange1:40101","exchange2:40102","exchange3:40103"})
func New(
	logger *slog.Logger,
	repo *repository.Repository,
	rdb *redis.Client,
	pairs []string,
	liveAddrs []string,
	resultCh chan Tick,
	initialMode Mode,
	db *sql.DB,
) *Service {
	priceSvc := NewPriceService(rdb, []string{"exchange1", "exchange2", "exchange3"})
	exSvc := NewExchangeService(rdb, pairs)
	srcMgr := NewSourceManager(exSvc, pairs, resultCh)
	modeSvc := NewModeService(logger, srcMgr, priceSvc, initialMode)
	health := NewHealthCheckService(db, rdb)

	agg := NewAggregator(rdb, repo.Aggregate, pairs, priceSvc)

	return &Service{
		HelthCheck:   health,
		ModeService:  modeSvc,
		PriceService: priceSvc,
		Aggregator:   agg, // ← добавили в возвращаемую структуру
	}
}
