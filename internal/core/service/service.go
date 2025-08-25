package service

import (
	"database/sql"

	"marketstream/internal/adapters/driven/database/repository"

	"github.com/redis/go-redis/v9"
)

type Service struct {
	ModeService  *ModeService
	PriceService *PriceService
	Exchange     *ExchangeService
	Aggregator   *Aggregator
	Sources      *SourceManager
	HelthCheck   *HealthCheckService
}

func New(repo *repository.Repository, rdb *redis.Client, pairs []string, exchanges []string, resultCh chan Tick, initialMode Mode, db *sql.DB) *Service {
	exSvc := NewExchangeService(rdb, pairs)
	srcMgr := NewSourceManager(exSvc, pairs, resultCh)
	modeSvc := NewModeService(srcMgr, initialMode)

	return &Service{
		ModeService:  modeSvc,
		PriceService: NewPriceService(rdb, exchanges),
		Exchange:     exSvc,
		Aggregator:   NewAggregator(rdb, repo.Aggregates, pairs, exchanges),
		Sources:      srcMgr,
		HelthCheck:   NewHealthCheckService(db, rdb),
	}
}
