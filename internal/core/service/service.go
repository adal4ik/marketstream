package service

import (
	"marketstream/internal/adapters/driven/database/repository"
	"marketstream/internal/adapters/driven/redis"
)

type Service struct {
	ModeService  *ModeService
	PriceService *PriceService
}

func New(repo repository.Repository, red *redis.Ouredis) *Service {
	return &Service{
		ModeService:  NewModeService(),
		PriceService: NewPriceService(repo.PriceRepository, red),
	}
}
