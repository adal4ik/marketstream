package exchange

import (
	"bufio"
	"context"
	"encoding/json"
	"log"
	"math"
	"net"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type ExchangeData struct {
	Symbol       string  `json:"symbol"`
	Price        float64 `json:"price"`
	Timestamp    int     `json:"timestamp"`
	ExchangeName string
}

type AggregateData struct {
	ExchangeName string
	Sum          float64
	Quantity     float64
	MinPrice     float64
	MaxPrice     float64
}

// Distributor
func ListenExchange(addr string, ctx context.Context, rdb *redis.Client) {
	conn, err := net.Dial("tcp", addr)
	exchangeName := strings.Split(addr, ":")[0]
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	dataChannel := make(chan []byte, 100)
	for i := 0; i < 5; i++ {
		go Worker(dataChannel, exchangeName, ctx, rdb)
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

// Worker
func Worker(dataChan <-chan []byte, exchangeName string, ctx context.Context, rdb *redis.Client) {
	for {
		select {
		case <-ctx.Done():
			return
		case data := <-dataChan:
			var processedData ExchangeData
			processedData.ExchangeName = exchangeName
			if err := json.Unmarshal(data, &processedData); err != nil {
				log.Fatalln("JSON ERROR: The new data couldn't be unmarshaled")
			}
			if err := rdb.Set(ctx, "latest:"+processedData.Symbol, processedData.Price, 60*time.Second).Err(); err != nil {
				log.Fatalln("REDIS ERROR: The new data couldn't be written into Redis")
			}
		}
	}
}

// Aggregator
func Aggregator(dataChan <-chan ExchangeData, ctx context.Context) {
	m := make(map[string]AggregateData)
	for {
		select {
		case <-ctx.Done():
			return
		case data := <-dataChan:
			var newAggregateData AggregateData
			newAggregateData.ExchangeName = data.ExchangeName
			newAggregateData.MaxPrice = math.Max(m[data.Symbol].MaxPrice, data.Price)
			newAggregateData.MinPrice = math.Min(m[data.Symbol].MinPrice, data.Price)
			newAggregateData.Quantity = m[data.Symbol].Quantity + 1
			newAggregateData.Sum = m[data.Symbol].Sum + data.Price
			m[data.Symbol] = newAggregateData
		}
	}
}
