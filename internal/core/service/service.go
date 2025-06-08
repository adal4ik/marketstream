package service

import (
	"marketstream/internal/adapters/driven/database/repository"

	"github.com/redis/go-redis/v9"
)

type Service struct {
	ModeService  *ModeService
	PriceService *PriceService
}

func New(repo repository.Repository, red *redis.Client) *Service {
	return &Service{
		ModeService:  NewModeService(),
		PriceService: NewPriceService(repo.PriceRepository, red),
	}
}
