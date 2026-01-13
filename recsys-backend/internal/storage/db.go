package storage

import (
	"context"
	"fmt"

	"recsys-backend/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewDB(ctx context.Context, cfg config.DBConfig) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name,
	)
	return pgxpool.New(ctx, dsn)
}
