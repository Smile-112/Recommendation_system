package storage

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed createDB.sql
var createDBSQL string

func EnsureSchema(ctx context.Context, db *pgxpool.Pool) error {
	var exists bool
	if err := db.QueryRow(ctx, `SELECT to_regclass('public."user"') IS NOT NULL`).Scan(&exists); err != nil {
		return fmt.Errorf("check schema: %w", err)
	}
	if exists {
		return nil
	}
	if _, err := db.Exec(ctx, createDBSQL); err != nil {
		return fmt.Errorf("init schema: %w", err)
	}
	return nil
}
