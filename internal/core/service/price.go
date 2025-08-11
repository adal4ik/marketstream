package service

import driveninterfaces "marketstream/internal/core/ports/driven"

type PriceService struct {
	red  driveninterfaces.RedisDrivenInterface
	repo driveninterfaces.PriceCache
}

func NewPriceService(repo driveninterfaces.PriceCache, red driveninterfaces.RedisDrivenInterface) *PriceService {
	return &PriceService{
		repo: repo,
		red:  red,
	}
}
