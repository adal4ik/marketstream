package domain

type Symbol string
type Exchange string

// Тик от биржи
type Tick struct {
	Exchange Exchange
	Symbol   Symbol
	Price    float64
	TsMilli  int64 // unix ms
}

func (t Tick) Valid() bool {
	return t.Price > 0 && t.TsMilli > 0 && t.Symbol != "" && t.Exchange != ""
}
