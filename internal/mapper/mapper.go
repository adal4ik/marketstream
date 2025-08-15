package mapper

import (
	"marketstream/internal/core/domain"
	"marketstream/internal/core/domain/dao"
)

func ToRow(d domain.MinuteAggregate) dao.MinuteAggregateRow {
	return dao.MinuteAggregateRow{
		PairName:     string(d.Pair),
		Exchange:     string(d.Exchange),
		Timestamp:    d.Ts.UTC(),
		AveragePrice: d.Avg,
		MinPrice:     d.Min,
		MaxPrice:     d.Max,
	}
}
