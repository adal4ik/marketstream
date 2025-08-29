package repository

import "database/sql"

type Repository struct {
	Aggregates *AggregateRepository
}

func New(db *sql.DB) *Repository {
	return &Repository{Aggregates: NewAggregateRepository(db)}
}
