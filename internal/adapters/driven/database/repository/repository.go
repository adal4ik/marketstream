package repository

import "database/sql"

type Repository struct {
	Price      *PriceRepository
	Aggregates *AggregateRepository
}

func New(db *sql.DB) *Repository {
	return &Repository{
		Price:      NewPriceRepository(db),
		Aggregates: NewAggregateRepository(db),
	}
}
