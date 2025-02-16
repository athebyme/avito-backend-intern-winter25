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
	err := row.Scan(&result.ID, &result.Name, &result.Price)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &domain.Merch{}, storage.ErrMerchNotFound
		}
		return &domain.Merch{}, fmt.Errorf("failed to find merch: %w", err)
	}

	return &result, nil
}

func (r *MerchRepository) GetAllAvailableMerch(ctx context.Context) ([]*domain.Merch, error) {
	query := `
        SELECT name, price
        FROM merch
    `
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var merch []*domain.Merch
	for rows.Next() {
		var t domain.Merch
		if err := rows.Scan(&t.Name, &t.Price); err != nil {
			return nil, err
		}
		merch = append(merch, &t)
	}
	return merch, nil
}
