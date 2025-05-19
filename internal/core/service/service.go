package service

type Service struct {
	ModeService  *ModeService
	PriceService *PriceService
}

func New() *Service {
	return &Service{
		ModeService:  NewModeService(),
		PriceService: NewPriceService(),
	}
}
