package entities

type PriceData struct {
	Pair      string  `json:"pair"`
	Price     float64 `json:"price"`
	Exchange  string  `json:"exchange"`
	Timestamp string  `json:"timestamp"`
}
