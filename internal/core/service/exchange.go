package service

import (
	"context"
	"encoding/json"
	"log"
	"marketstream/internal/core/entities"
	driveninterfaces "marketstream/internal/core/interfaces/driven"
	"time"
)

type ExchangeService struct {
	rdb driveninterfaces.RedisDrivenInterface
	db  driveninterfaces.DataBaseInterface
}

func NewExchangeService(db driveninterfaces.DataBaseInterface, rdb driveninterfaces.RedisDrivenInterface) *ExchangeService {
	return &ExchangeService{
		rdb: rdb,
		db:  db,
	}
}

func (exchange *ExchangeService) Worker(dataChan <-chan []byte, exchangeName string, ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case data := <-dataChan:
			var processedData entities.ExchangeData
			processedData.ExchangeName = exchangeName
			if err := json.Unmarshal(data, &processedData); err != nil {
				log.Fatalln("JSON ERROR: The new data couldn't be unmarshaled")
			}
			if err := exchange.rdb.Set(ctx, "latest:"+processedData.Symbol, processedData.Price, 60*time.Second); err != nil {
				log.Fatalln("REDIS ERROR: The new data couldn't be written into Redis")
			}
		}
	}
}
