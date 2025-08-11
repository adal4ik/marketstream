package service

import (
	"marketstream/internal/adapters/driven/database/repository"

	"github.com/redis/go-redis/v9"
)

type Service struct {
	ModeService  *ModeService
	PriceService *PriceService
	Exchange     *ExchangeService // ← добавили поле
}

// Добавили pairs, чтобы ExchangeService знал валидные пары
func New(repo *repository.Repository, red *redis.Client, pairs []string) *Service {
	return &Service{
		ModeService: NewModeService(),
		// PriceService: NewPriceService(repo, red),
		Exchange: NewExchangeService(red, pairs), // ← инициализируем
	}
}
