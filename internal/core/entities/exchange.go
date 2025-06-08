package entities

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
