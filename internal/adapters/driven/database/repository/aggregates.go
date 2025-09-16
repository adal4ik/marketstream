package repository

import (
	"context"
	"database/sql"
	"time"
)

type MinuteAggregateRecord struct {
	PairName     string
	Exchange     string
	Timestamp    time.Time
	AveragePrice float64
	MinPrice     float64
	MaxPrice     float64
}

type AggregateRepository struct {
	db *sql.DB
}

// BatchInsert: upsert минутных агрегатов.
// Если запись уже есть за ту же минуту — обновляем avg, min, max.
func (r *AggregateRepository) BatchInsert(ctx context.Context, recs []MinuteAggregateRecord) error {
	if len(recs) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO minute_aggregates
			(pair_name, exchange, "timestamp", average_price, min_price, max_price)
		VALUES
			($1,$2,$3,$4,$5,$6)
		ON CONFLICT (pair_name, exchange, "timestamp")
		DO UPDATE SET
			average_price = EXCLUDED.average_price,
			min_price     = LEAST(minute_aggregates.min_price, EXCLUDED.min_price),
			max_price     = GREATEST(minute_aggregates.max_price, EXCLUDED.max_price)
	`)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	defer stmt.Close()

	for _, rec := range recs {
		if _, err := stmt.ExecContext(
			ctx,
			rec.PairName,
			rec.Exchange,
			rec.Timestamp, // TIMESTAMPTZ
			rec.AveragePrice,
			rec.MinPrice,
			rec.MaxPrice,
		); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}
