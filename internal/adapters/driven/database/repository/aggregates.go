package repository

import (
	"context"
	"database/sql"
	"marketstream/internal/core/domain"
	"marketstream/internal/mapper"
	"strconv"
	"strings"
)

type AggregateRepository struct {
	db *sql.DB
}

func NewAggregateRepository(db *sql.DB) *AggregateRepository {
	return &AggregateRepository{
		db: db,
	}
}

func (r *AggregateRepository) BatchInsert(ctx context.Context, recs []domain.MinuteAggregate) error {
	if len(recs) == 0 {
		return nil
	}
	var sb strings.Builder
	args := make([]any, 0, len(recs)*6)

	sb.WriteString(`INSERT INTO minute_aggregates (pair_name, exchange, "timestamp", average_price, min_price, max_price) VALUES `)

	for i, d := range recs {
		if i > 0 {
			sb.WriteString(",")
		}
		base := i*6 + 1
		sb.WriteString("(" +
			"$" + strconv.Itoa(base+0) + "," +
			"$" + strconv.Itoa(base+1) + "," +
			"$" + strconv.Itoa(base+2) + "," +
			"$" + strconv.Itoa(base+3) + "," +
			"$" + strconv.Itoa(base+4) + "," +
			"$" + strconv.Itoa(base+5) + ")")

		row := mapper.ToRow(d)
		args = append(args, row.PairName, row.Exchange, row.Timestamp, row.AveragePrice, row.MinPrice, row.MaxPrice)
	}

	_, err := r.db.ExecContext(ctx, sb.String(), args...)
	return err
}
