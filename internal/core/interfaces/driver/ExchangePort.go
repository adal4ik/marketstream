package driverinterfaces

import (
	"context"
)

type ExchangePortInterface interface {
	Worker(dataChan <-chan []byte, exchangeName string, ctx context.Context)
}
