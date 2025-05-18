package entities

type Pairs struct {
	PairID       string  `json:"pair_id"`
	PairName     string  `json:"pair_name"`
	Exchange     string  `json:"exchange"`
	AveragePrice float64 `json:"average_price"`
	MinPrice     float64 `json:"min_price"`
	MaxPrice     float64 `json:"max_price"`
}
