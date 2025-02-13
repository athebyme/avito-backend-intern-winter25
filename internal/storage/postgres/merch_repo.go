package postgres

import (
	"avito-backend-intern-winter25/internal/models"
	"avito-backend-intern-winter25/internal/storage"
	"database/sql"
)

type MerchRepository struct {
	db *sql.DB
}

func NewMerchRepository(db *sql.DB) *MerchRepository {
	return &MerchRepository{
		db: db,
	}
}

func (r *MerchRepository) FindByName(name string) (models.Merch, error) {
	if name == "" {
		return models.Merch{}, storage.ErrMerchNameIsIncorrect
	}
	query := `
	SELECT id, name, price FROM merch WHERE name = $1;
	`

	var result models.Merch

	row := r.db.QueryRow(query, name)
	err := row.Scan(&result.Id, &result.Name, &result.Price)
	if err != nil {
		return models.Merch{}, storage.ErrMerchNotFound
	}

	return result, nil
}
