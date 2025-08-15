package domain

import "time"

type MinuteAggregate struct {
	Pair     Symbol
	Exchange Exchange
	Ts       time.Time // UTC
	Avg      float64
	Min      float64
	Max      float64
}
