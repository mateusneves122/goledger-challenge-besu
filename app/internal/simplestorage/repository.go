package simplestorage

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type PostgresRepo struct {
	db *sql.DB
}

func NewPostgresRepo(db *sql.DB) *PostgresRepo {
	return &PostgresRepo{db: db}
}

func (r *PostgresRepo) Save(ctx context.Context, contractAddress, value string) error {
	saveCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := r.db.ExecContext(saveCtx,
		`INSERT INTO contract_values (contract_address, value, synced_at)
		 VALUES ($1, $2, NOW())
		 ON CONFLICT (contract_address) DO UPDATE
		     SET value = EXCLUDED.value, synced_at = NOW()`,
		contractAddress, value,
	)
	if err != nil {
		return fmt.Errorf("upsert contract value: %w", err)
	}
	return nil
}

func (r *PostgresRepo) GetLatest(ctx context.Context, contractAddress string) (string, error) {
	getCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var value string
	err := r.db.QueryRowContext(getCtx,
		`SELECT value FROM contract_values WHERE contract_address = $1`,
		contractAddress,
	).Scan(&value)

	if err == sql.ErrNoRows {
		return "0", nil
	}
	if err != nil {
		return "", fmt.Errorf("query contract value: %w", err)
	}

	return value, nil
}
