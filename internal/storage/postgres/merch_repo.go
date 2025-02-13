package postgres

import (
	"avito-backend-intern-winter25/internal/models/domain"
	"avito-backend-intern-winter25/internal/storage"
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type MerchRepository struct {
	db *sql.DB
}

func NewMerchRepository(db *sql.DB) *MerchRepository {
	return &MerchRepository{
		db: db,
	}
}

func (r *MerchRepository) FindByName(ctx context.Context, name string) (*domain.Merch, error) {
	if name == "" {
		return &domain.Merch{}, storage.ErrMerchNameIsIncorrect
	}
	query := `
	SELECT id, name, price FROM merch WHERE name = $1;
	`

	var result domain.Merch

	row := r.db.QueryRowContext(ctx, query, name)
	err := row.Scan(&result.Id, &result.Name, &result.Price)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &domain.Merch{}, storage.ErrMerchNotFound
		}
		return &domain.Merch{}, fmt.Errorf("failed to find merch: %w", err)
	}

	return &result, nil
}
