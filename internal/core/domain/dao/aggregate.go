package dao

import "time"

// Соответствует таблице minute_aggregates
type MinuteAggregateRow struct {
	PairName     string
	Exchange     string
	Timestamp    time.Time
	AveragePrice float64
	MinPrice     float64
	MaxPrice     float64
}
