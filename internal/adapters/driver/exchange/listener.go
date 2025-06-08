package exchange

import (
	"bufio"
	"context"
	"log"
	driverinterfaces "marketstream/internal/core/interfaces/driver"
	"net"
	"strings"
)

type ExchangeHandler struct {
	exchangeService driverinterfaces.ExchangePortInterface
}

func NewExchageHandler(exchangeservice driverinterfaces.ExchangePortInterface) *ExchangeHandler {
	return &ExchangeHandler{
		exchangeService: exchangeservice,
	}
}

// Distributor
func (exchange *ExchangeHandler) ListenExchange(addr string, ctx context.Context) {
	conn, err := net.Dial("tcp", addr)
	exchangeName := strings.Split(addr, ":")[0]
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	dataChannel := make(chan []byte, 100)
	for i := 0; i < 5; i++ {
		go exchange.exchangeService.Worker(dataChannel, exchangeName, ctx)
	}

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		newData := scanner.Bytes()
		dataChannel <- newData
	}

	if err := scanner.Err(); err != nil {
		log.Println(err)
	}
}

// // Aggregator
// func Aggregator(dataChan <-chan entities.ExchangeData, ctx context.Context) {
// 	m := make(map[string]entities.AggregateData)
// 	for {
// 		select {
// 		case <-ctx.Done():
// 			return
// 		case data := <-dataChan:
// 			var newAggregateData entities.AggregateData
// 			newAggregateData.ExchangeName = data.ExchangeName
// 			newAggregateData.MaxPrice = math.Max(m[data.Symbol].MaxPrice, data.Price)
// 			newAggregateData.MinPrice = math.Min(m[data.Symbol].MinPrice, data.Price)
// 			newAggregateData.Quantity = m[data.Symbol].Quantity + 1
// 			newAggregateData.Sum = m[data.Symbol].Sum + data.Price
// 			m[data.Symbol] = newAggregateData
// 		}
// 	}
// }
