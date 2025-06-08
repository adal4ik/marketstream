package service

import driveninterfaces "marketstream/internal/core/interfaces/driven"

type PriceService struct {
	red  driveninterfaces.RedisDrivenInterface
	repo driveninterfaces.PriceDrivenInterface
}

func NewPriceService(repo driveninterfaces.PriceDrivenInterface, red driveninterfaces.RedisDrivenInterface) *PriceService {
	return &PriceService{
		repo: repo,
		red:  red,
	}
}
