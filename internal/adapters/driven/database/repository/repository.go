package repository

import "database/sql"

type Repository struct {
	PriceRepository *PriceRepository
}

func New(db *sql.DB) *Repository {
	return &Repository{
		PriceRepository: NewPriceRepository(db),
	}
}
