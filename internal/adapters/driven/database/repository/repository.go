package repository

import "database/sql"

type Repository struct {
	DB        *sql.DB
	Aggregate *AggregateRepository
}

func New(db *sql.DB) *Repository {
	return &Repository{
		DB:        db,
		Aggregate: &AggregateRepository{db: db},
	}
}
