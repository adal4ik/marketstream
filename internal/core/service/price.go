package service

import (
	driveninterfaces "marketstream/internal/core/interfaces/driven"
)

type PriceService struct {
	red  driveninterfaces.RedisDrivenInterface
	repo driveninterfaces.DataBaseInterface
}

func NewPriceService(repo driveninterfaces.DataBaseInterface, red driveninterfaces.RedisDrivenInterface) *PriceService {
	return &PriceService{
		repo: repo,
		red:  red,
	}
}
