package drivenports

import (
	"context"

	"marketstream/internal/core/domain"
)

type AggregateRepo interface {
	BatchInsert(ctx context.Context, recs []domain.MinuteAggregate) error
}
